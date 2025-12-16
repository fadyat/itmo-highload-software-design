package shell

import (
	"errors"
	"fmt"
	"regexp"
)

// Arg keeps token text and quote information for later substitution/handling.
type Arg struct {
	Text  string
	Quote rune // 0 = none, '\'' = single, '"' = double
}

// Command represents a single command in a pipeline. If Name is empty and
// Assignments is non-empty, this command is treated as variable assignments.
type Command struct {
	Name        string
	Args        []Arg
	Assignments map[string]string
}

// Pipeline is a sequence of commands separated by pipes.
type Pipeline struct {
	Commands []Command
}

// ErrInvalidSyntax returned on simple parse errors.
var ErrInvalidSyntax = errors.New("invalid syntax")

var varNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Parse converts a slice of Tokens (as produced by Lex) into a Pipeline of Commands.
// It recognizes:
//   - pipe tokens (Token.Quote == '|') as command separators
//   - assignments of the form NAME = value (lexer produces '=' as a separate token with Quote=='=')
//     When a sequence NAME '=' VALUE is encountered it is recorded in Command.Assignments.
//     Multiple assignments may appear before a command name; if no command name follows they are
//     treated as standalone assignments (command with empty Name).
//
// The parser preserves quotes on arguments (in Arg.Quote) so later stages (executor / env manager)
// can perform substitutions according to quoting rules.
func Parse(tokens []Token) (Pipeline, error) {
	var p Pipeline

	// quick handle empty input
	if len(tokens) == 0 {
		return p, nil
	}

	i := 0
	n := len(tokens)

	// helper to check assignment start at position i: NAME '=' VALUE
	isAssignmentStart := func(i int) bool {
		// need at least 3 tokens: NAME, '=', VALUE
		if i+2 >= n {
			return false
		}
		// '=' token produced by lexer has Quote == '='
		if tokens[i+1].Quote != '=' {
			return false
		}
		// NAME must be unquoted and valid identifier
		if tokens[i].Quote != 0 {
			return false
		}
		if !varNameRE.MatchString(tokens[i].Text) {
			return false
		}
		// VALUE can be quoted or unquoted; accept any token
		return true
	}

	for i < n {
		// reject pipe at start
		if tokens[i].Quote == '|' {
			return p, fmt.Errorf("%w: unexpected pipe at position %d", ErrInvalidSyntax, i)
		}

		cmd := Command{
			Assignments: make(map[string]string),
			Args:        nil,
			Name:        "",
		}

		// collect tokens for this command until pipe or end
		for i < n && tokens[i].Quote != '|' {
			// assignment handling
			if isAssignmentStart(i) {
				name := tokens[i].Text
				valTok := tokens[i+2]
				cmd.Assignments[name] = valTok.Text
				i += 3
				// allow multiple assignments in a row
				continue
			}

			// Otherwise treat as command name or argument
			// If command name is not set yet, set it to token text.
			// Note: even if token was quoted, we use the Text (quotes removed by lexer).
			if cmd.Name == "" {
				cmd.Name = tokens[i].Text
			} else {
				cmd.Args = append(cmd.Args, Arg{
					Text:  tokens[i].Text,
					Quote: tokens[i].Quote,
				})
			}
			i++
		}

		// append command (even if Name empty — it may be assignment-only)
		p.Commands = append(p.Commands, cmd)

		// skip pipe separator if present
		if i < n && tokens[i].Quote == '|' {
			// reject trailing pipe (pipe at end)
			if i == n-1 {
				return p, fmt.Errorf("%w: trailing pipe", ErrInvalidSyntax)
			}
			i++ // skip '|'
		}
	}

	return p, nil
}

// ParseString is a convenience: lex the input and parse into a Pipeline.
func ParseString(input string) (Pipeline, error) {
	toks := Lex(input)
	return Parse(toks)
}
