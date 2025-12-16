package shell

import (
	"strings"
	"unicode"
)

// Token represents a lexed token with quoting information.
// Text: the token content with surrounding quotes removed (if any).
// Quote: 0 = none, '\” = single quote, '"' = double quote, '|' = pipe token, '=' = equal token
type Token struct {
	Text  string
	Quote rune
}

// Lex splits input into tokens, respecting single and double quotes and recognizing
// special single-character tokens: pipe '|' and equal '='.
//
// Rules implemented:
//   - Whitespace (unicode.IsSpace) separates tokens.
//   - '|' and '=' are always treated as separate tokens.
//   - Single quotes (”) — full quoting: everything inside is taken literally (no escapes).
//   - Double quotes ("") — weak quoting: everything inside is taken literally for lexing purposes,
//     but later stages (environment substitution) may replace $VAR occurrences. Backslashes are NOT processed here.
//   - Unquoted tokens may contain backslash-escaped characters: a backslash causes the next rune
//     to be included verbatim (i.e. it's removed and the rune is kept).
//
// The returned tokens have their surrounding quotes removed; for quoted tokens the Quote
// field is set to the quote rune so downstream code can decide how to treat them.
func Lex(input string) []Token {
	var tokens []Token
	runes := []rune(input)
	n := len(runes)
	for i := 0; i < n; {
		// skip whitespace
		if unicode.IsSpace(runes[i]) {
			i++
			continue
		}

		// special single-char tokens
		if runes[i] == '|' {
			tokens = append(tokens, Token{Text: "|", Quote: '|'})
			i++
			continue
		}
		if runes[i] == '=' {
			tokens = append(tokens, Token{Text: "=", Quote: '='})
			i++
			continue
		}

		// quoted tokens
		if runes[i] == '\'' || runes[i] == '"' {
			q := runes[i]
			i++
			var sb strings.Builder
			for i < n && runes[i] != q {
				// For single quotes: take everything as-is.
				// For double quotes: also take everything as-is for lexing; substitution happens later.
				sb.WriteRune(runes[i])
				i++
			}
			// If we stopped at a closing quote, skip it
			if i < n && runes[i] == q {
				i++
			}
			tokens = append(tokens, Token{Text: sb.String(), Quote: q})
			continue
		}

		// unquoted token
		var sb strings.Builder
		for i < n && !unicode.IsSpace(runes[i]) && runes[i] != '|' && runes[i] != '=' {
			if runes[i] == '\\' {
				// escape the next rune if present
				i++
				if i < n {
					sb.WriteRune(runes[i])
					i++
				}
				continue
			}
			sb.WriteRune(runes[i])
			i++
		}
		tokens = append(tokens, Token{Text: sb.String(), Quote: 0})
	}
	return tokens
}
