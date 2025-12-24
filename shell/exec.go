package shell

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
)

// Env manages environment variables for the shell.
type Env struct {
	vars map[string]string
	mu   sync.RWMutex
}

// NewEnvFromOS initializes Env with current OS environment variables.
func NewEnvFromOS() *Env {
	e := &Env{vars: make(map[string]string)}
	for _, kv := range os.Environ() {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) == 2 {
			e.vars[parts[0]] = parts[1]
		}
	}
	return e
}

// Get returns variable value or empty string if not present.
func (e *Env) Get(key string) string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.vars[key]
}

// Set assigns a variable.
func (e *Env) Set(key, val string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.vars[key] = val
}

// Clone returns a shallow copy of the environment (copy of the vars map).
// The returned Env has its own map so modifications to it do not affect the original.
func (e *Env) Clone() *Env {
	e.mu.RLock()
	defer e.mu.RUnlock()
	newMap := make(map[string]string, len(e.vars))
	for k, v := range e.vars {
		newMap[k] = v
	}
	return &Env{vars: newMap}
}

// Substitute performs variable substitution on input string according to quoting rules.
// If quote == '\”, substitution is disabled (single quotes).
// Otherwise, occurrences of $NAME are replaced with variable values (or empty if missing).
var varRefRE = regexp.MustCompile(`\$(?:\{([A-Za-z_][A-Za-z0-9_]*)\}|([A-Za-z_][A-Za-z0-9_]*))`)

func (e *Env) Substitute(s string, quote rune) string {
	if quote == '\'' {
		return s
	}
	// Replace $VAR and ${VAR}
	return varRefRE.ReplaceAllStringFunc(s, func(m string) string {
		// extract name
		var name string
		if strings.HasPrefix(m, "${") && strings.HasSuffix(m, "}") {
			name = m[2 : len(m)-1]
		} else if strings.HasPrefix(m, "$") {
			name = m[1:]
		}
		if name == "" {
			return ""
		}
		val := e.Get(name)
		return val
	})
}

// Executor errors
var ErrExitRequested = errors.New("exit requested")

// ExecutePipeline executes a Pipeline using provided Env.
// in/out/errOut default to os.Stdin/os.Stdout/os.Stderr if nil.
// Returns exit code of the last command (or exit code requested by builtin exit) and error.
// If a builtin `exit` is executed, returns ErrExitRequested and the requested code as the int.
func ExecutePipeline(p Pipeline, env *Env, in io.Reader, out io.Writer, errOut io.Writer) (int, error) {
	if env == nil {
		env = NewEnvFromOS()
	}
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}

	// Apply any top-level assignments before running commands.
	// Also handle commands that are assignment-only.
	if len(p.Commands) == 0 {
		return 0, nil
	}

	// If pipeline has a single command that's only assignments, apply them and return.
	if len(p.Commands) == 1 && p.Commands[0].Name == "" && len(p.Commands[0].Assignments) > 0 {
		for k, v := range p.Commands[0].Assignments {
			env.Set(k, v)
		}
		return 0, nil
	}

	// Build pipeline I/O:
	n := len(p.Commands)
	// readers[i] is stdin for command i
	readers := make([]io.Reader, n)
	// writers[i] is stdout for command i
	writers := make([]io.Writer, n)

	readers[0] = in
	for i := 0; i < n; i++ {
		if i == n-1 {
			writers[i] = out
		} else {
			r, w := io.Pipe()
			writers[i] = w
			readers[i+1] = r
		}
	}

	// Run each command concurrently (so pipes work); collect exit codes
	type result struct {
		idx  int
		code int
		err  error
	}
	resCh := make(chan result, n)
	var wg sync.WaitGroup
	wg.Add(n)

	// Compute per-command environments for pipeline-local assignments.
	// We treat the pipeline as having its own local environment snapshot,
	// initialized from the provided env. Assignment-only commands inside the
	// pipeline modify this local snapshot (do not mutate the global env).
	// Each actual command executes with a clone of the current pipeline snapshot
	// at the moment it appears; if the command itself has assignments those are
	// applied to that clone only.
	usedEnvs := make([]*Env, n)
	pipelineEnv := env.Clone() // pipeline-local snapshot
	for i := 0; i < n; i++ {
		cmd := p.Commands[i]
		// If assignment-only command: apply to pipeline-local snapshot
		if cmd.Name == "" && len(cmd.Assignments) > 0 {
			for k, v := range cmd.Assignments {
				pipelineEnv.Set(k, v)
			}
			// usedEnvs[i] remains a clone of the current pipelineEnv for completeness
			usedEnvs[i] = pipelineEnv.Clone()
			continue
		}
		// For a real command, use a clone of the current pipeline snapshot
		// so concurrent execution is safe. Apply command-local assignments
		// to that clone only if present.
		local := pipelineEnv.Clone()
		if len(cmd.Assignments) > 0 {
			for k, v := range cmd.Assignments {
				local.Set(k, v)
			}
		}
		usedEnvs[i] = local
	}

	for i := 0; i < n; i++ {
		ci := p.Commands[i]
		stdin := readers[i]
		stdout := writers[i]
		stderr := errOut
		usedEnv := usedEnvs[i]

		go func(cmd Command, r io.Reader, w io.Writer, eout io.Writer, idx int, envForCmd *Env) {
			defer wg.Done()

			// If command has no name (assignment-only command within pipeline),
			// we've already applied assignments to the pipeline-local snapshot; produce no output and succeed.
			if cmd.Name == "" {
				// close writer if it's a pipe writer so next command gets EOF
				if pw, ok := w.(*io.PipeWriter); ok {
					_ = pw.Close()
				}
				resCh <- result{idx: idx, code: 0, err: nil}
				return
			}

			// Build argument strings with substitution using envForCmd
			args := make([]string, 0, len(cmd.Args))
			for _, a := range cmd.Args {
				val := envForCmd.Substitute(a.Text, a.Quote)
				args = append(args, val)
			}

			// Substitution of command name itself (rare) - treat as unquoted, using envForCmd.
			name := envForCmd.Substitute(cmd.Name, 0)

			// Builtins handling (simplified)
			if isBuiltin(name) {
				code, err := runBuiltin(name, args, r, w, eout)
				// Close this command's stdin pipe reader (if any) when the builtin finishes,
				// so upstream writers won't block trying to write into this now-unused pipe.
				// Previously we only closed the reader for ErrExitRequested; close it
				// unconditionally here to avoid deadlocks when the builtin returns early
				// (for example, due to argument parse errors) and the upstream writer is still active.
				if pr, ok := r.(*io.PipeReader); ok {
					_ = pr.Close()
				}
				// close pipe writer if necessary
				closeIfPipe(w)
				resCh <- result{idx: idx, code: code, err: err}
				return
			}

			// External command execution (pass envForCmd so external process sees temporary vars)
			exitCode, err := runExternal(name, args, r, w, eout, envForCmd)
			closeIfPipe(w)
			resCh <- result{idx: idx, code: exitCode, err: err}
		}(ci, stdin, stdout, stderr, i, usedEnv)
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resCh)

	// Collect results - find last command's result
	finalCode := 0
	var finalErr error
	for res := range resCh {
		if res.idx == n-1 {
			finalCode = res.code
			finalErr = res.err
		}
		// If any command requested exit, propagate it (give priority to first encountered)
		if res.err == ErrExitRequested {
			// when exit is requested, choose that code as the final code
			return res.code, ErrExitRequested
		}
	}
	return finalCode, finalErr
}

// runExternal executes an external program with given stdin/stdout/stderr and environment substitution.
// Helper to identify builtin commands
func closeIfPipe(w io.Writer) {
	if pw, ok := w.(*io.PipeWriter); ok {
		_ = pw.Close()
	}
}

func isBuiltin(name string) bool {
	switch name {
	case "echo", "cat", "wc", "pwd", "grep", "exit":
		return true
	}
	return false
}

// runBuiltin dispatches to builtin implementations. For `exit` it returns ErrExitRequested as error and provides the code.
func runBuiltin(name string, args []string, r io.Reader, w io.Writer, eout io.Writer) (int, error) {
	switch name {
	case "echo":
		return builtinEcho(args, r, w, eout)
	case "cat":
		return builtinCat(args, r, w, eout)
	case "wc":
		return builtinWc(args, r, w, eout)
	case "pwd":
		return builtinPwd(r, w, eout)
	case "grep":
		return builtinGrep(args, r, w, eout)
	case "exit":
		code := 0
		if len(args) > 0 {
			if n, err := strconv.Atoi(args[0]); err == nil {
				code = n
			}
		}
		return code, ErrExitRequested
	}
	return 1, nil
}

// runExternal executes an external program with given stdin/stdout/stderr.
// It constructs environment variables from os.Env plus our Env.vars (our env overrides OS env).
func runExternal(name string, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer, env *Env) (int, error) {
	// look up executable in PATH (use exec.Command which does that)
	cmd := exec.Command(name, args...)
	// Wire IO
	if stdin != nil {
		cmd.Stdin = stdin
	}
	if stdout != nil {
		cmd.Stdout = stdout
	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	// Build environment
	baseEnv := os.Environ()
	// apply overrides from env
	env.mu.RLock()
	for k, v := range env.vars {
		baseEnv = append(baseEnv, k+"="+v)
	}
	env.mu.RUnlock()
	cmd.Env = baseEnv

	if err := cmd.Start(); err != nil {
		return 1, err
	}
	if err := cmd.Wait(); err != nil {
		// Try to extract exit code (best-effort)
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status := exitErr.ProcessState; status != nil {
				// no portable way to get code across systems without syscall
				// best-effort: return 1
				_ = status
			}
		}
		return 1, err
	}
	return 0, nil
}

// Builtin implementations

func builtinEcho(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	// echo writes args joined by spaces and newline
	if stdout == nil {
		stdout = os.Stdout
	}
	_, err := io.WriteString(stdout, strings.Join(args, " ")+"\n")
	if err != nil {
		return 1, err
	}
	return 0, nil
}

func builtinCat(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	// If args provided, cat each file in order
	if len(args) == 0 {
		_, err := io.Copy(stdout, stdin)
		if err != nil {
			return 1, err
		}
		return 0, nil
	}
	for _, name := range args {
		// Support '-' as stdin
		if name == "-" {
			if _, err := io.Copy(stdout, stdin); err != nil {
				return 1, err
			}
			continue
		}
		// Open file relative to cwd
		f, err := os.Open(filepath.Clean(name))
		if err != nil {
			if stderr != nil {
				_, _ = io.WriteString(stderr, err.Error()+"\n")
			}
			return 1, err
		}
		_, err = io.Copy(stdout, f)
		_ = f.Close()
		if err != nil {
			if stderr != nil {
				_, _ = io.WriteString(stderr, err.Error()+"\n")
			}
			return 1, err
		}
	}
	return 0, nil
}

func builtinWc(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	// helper to process a reader and return counts
	process := func(r io.Reader) (lines, words, byteCount int, err error) {
		var buf bytes.Buffer
		n, err := buf.ReadFrom(r)
		if err != nil {
			return 0, 0, 0, err
		}
		b := buf.Bytes()
		byteCount = int(n)
		lines = bytesCount(b, '\n')
		// words via scanner
		sc := bufio.NewScanner(bytes.NewReader(b))
		sc.Split(bufio.ScanWords)
		for sc.Scan() {
			words++
		}
		// ignore scanner error for simplicity
		return lines, words, byteCount, nil
	}

	if len(args) == 0 {
		l, w, b, err := process(stdin)
		if err != nil {
			return 1, err
		}
		_, _ = io.WriteString(stdout, strconv.Itoa(l)+" "+strconv.Itoa(w)+" "+strconv.Itoa(b)+"\n")
		return 0, nil
	}

	// multiple files
	totalL, totalW, totalB := 0, 0, 0
	for _, name := range args {
		if name == "-" {
			l, w, b, err := process(stdin)
			if err != nil {
				if stderr != nil {
					_, _ = io.WriteString(stderr, err.Error()+"\n")
				}
				return 1, err
			}
			_, _ = io.WriteString(stdout, strconv.Itoa(l)+" "+strconv.Itoa(w)+" "+strconv.Itoa(b)+"\n")
			totalL += l
			totalW += w
			totalB += b
			continue
		}
		f, err := os.Open(filepath.Clean(name))
		if err != nil {
			if stderr != nil {
				_, _ = io.WriteString(stderr, err.Error()+"\n")
			}
			return 1, err
		}
		l, w, b, err := process(f)
		_ = f.Close()
		if err != nil {
			if stderr != nil {
				_, _ = io.WriteString(stderr, err.Error()+"\n")
			}
			return 1, err
		}
		_, _ = io.WriteString(stdout, strconv.Itoa(l)+" "+strconv.Itoa(w)+" "+strconv.Itoa(b)+" "+name+"\n")
		totalL += l
		totalW += w
		totalB += b
	}
	if len(args) > 1 {
		_, _ = io.WriteString(stdout, strconv.Itoa(totalL)+" "+strconv.Itoa(totalW)+" "+strconv.Itoa(totalB)+" total\n")
	}
	return 0, nil
}

func bytesCount(b []byte, sep byte) int {
	if len(b) == 0 {
		return 0
	}
	c := 0
	for _, x := range b {
		if x == sep {
			c++
		}
	}
	return c
}

func builtinPwd(stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	dir, err := os.Getwd()
	if err != nil {
		if stderr != nil {
			_, _ = io.WriteString(stderr, err.Error()+"\n")
		}
		return 1, err
	}
	_, _ = io.WriteString(stdout, dir+"\n")
	return 0, nil
}

// builtinGrep implements a small subset of GNU grep features required by the assignment.
// Supported flags:
//
//	-w : match whole words (word boundaries)
//	-i : case-insensitive matching
//	-A N : print N lines After a matching line (context)
//
// Pattern is a regular expression. If files are provided, each is searched; otherwise stdin is used.
// When multiple files are searched, printed lines are prefixed with "filename:".
func builtinGrep(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (int, error) {
	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}

	// Define options for go-flags parser.
	var opts struct {
		Word       bool `short:"w" long:"word" description:"match whole words only"`
		IgnoreCase bool `short:"i" long:"ignore-case" description:"ignore case distinctions"`
		After      int  `short:"A" long:"after" description:"print NUM lines of trailing context" default:"0"`
	}

	// Parse args using go-flags to satisfy the requirement to use a CLI args library.
	parser := flags.NewParser(&opts, flags.IgnoreUnknown)
	remaining, err := parser.ParseArgs(args)
	if err != nil {
		// write parser error to stderr and return non-zero exit
		_, _ = io.WriteString(stderr, "grep: "+err.Error()+"\n")
		return 2, err
	}

	if len(remaining) == 0 {
		_, _ = io.WriteString(stderr, "grep: missing search pattern\n")
		return 2, nil
	}

	pattern := remaining[0]
	files := remaining[1:]

	// If -w requested, wrap pattern with word boundaries. Note: \b in Go's regexp follows RE2 semantics.
	if opts.Word {
		pattern = `\b` + pattern + `\b`
	}

	// If case-insensitive, prefix pattern with (?i)
	if opts.IgnoreCase {
		pattern = "(?i)" + pattern
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		_, _ = io.WriteString(stderr, "grep: invalid regexp: "+err.Error()+"\n")
		return 2, err
	}

	// Helper to process a reader line-by-line and write matching lines to stdout.
	processReader := func(r io.Reader, name string, prefixWithName bool) error {
		sc := bufio.NewScanner(r)
		// default Scanner buffer is usually sufficient for typical lines; this is a simple implementation.
		afterRem := 0
		for sc.Scan() {
			line := sc.Text()
			matched := re.MatchString(line)
			if matched {
				// print this line
				if prefixWithName {
					_, _ = io.WriteString(stdout, name+":"+line+"\n")
				} else {
					_, _ = io.WriteString(stdout, line+"\n")
				}
				afterRem = opts.After
				continue
			}
			// if within trailing context -A, print the line
			if afterRem > 0 {
				if prefixWithName {
					_, _ = io.WriteString(stdout, name+":"+line+"\n")
				} else {
					_, _ = io.WriteString(stdout, line+"\n")
				}
				afterRem--
				continue
			}
			// otherwise skip
		}
		// If scanner experienced an error, report it
		if err := sc.Err(); err != nil {
			return err
		}
		return nil
	}

	// If files provided, iterate them; otherwise read from stdin
	if len(files) > 0 {
		prefix := len(files) > 1
		for _, fname := range files {
			// support '-' as stdin
			if fname == "-" {
				if err := processReader(stdin, "<stdin>", prefix); err != nil {
					_, _ = io.WriteString(stderr, "grep: error reading stdin: "+err.Error()+"\n")
					return 1, err
				}
				continue
			}
			f, err := os.Open(filepath.Clean(fname))
			if err != nil {
				_, _ = io.WriteString(stderr, "grep: "+err.Error()+"\n")
				// continue to next file (similar to grep behaviour which may report and continue)
				continue
			}
			if err := processReader(f, fname, prefix); err != nil {
				_ = f.Close()
				_, _ = io.WriteString(stderr, "grep: error reading "+fname+": "+err.Error()+"\n")
				return 1, err
			}
			_ = f.Close()
		}
	} else {
		// No files: read from stdin
		if stdin == nil {
			stdin = os.Stdin
		}
		if err := processReader(stdin, "<stdin>", false); err != nil {
			_, _ = io.WriteString(stderr, "grep: error reading stdin: "+err.Error()+"\n")
			return 1, err
		}
	}

	return 0, nil
}
