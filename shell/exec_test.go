package shell

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// runPipeline is a small helper: parse input and execute pipeline capturing stdout.
func runPipeline(input string, env *Env) (string, int, error) {
	p, err := ParseString(input)
	if err != nil {
		return "", 0, err
	}
	var out bytes.Buffer
	var errOut bytes.Buffer
	code, err := ExecutePipeline(p, env, nil, &out, &errOut)
	// prefer returning stderr content in error message if execution failed
	if err != nil && err != ErrExitRequested {
		return out.String() + errOut.String(), code, err
	}
	return out.String(), code, err
}

func TestLexBasic(t *testing.T) {
	in := `echo "Hello $USER" | wc`
	got := Lex(in)
	if len(got) != 4 {
		t.Fatalf("unexpected tokens: %#v", got)
	}
	if got[0].Text != "echo" || got[1].Text != "Hello $USER" || got[2].Text != "|" || got[3].Text != "wc" {
		t.Fatalf("lex produced wrong tokens: %#v", got)
	}
}

func TestParseAssignmentsAndArgs(t *testing.T) {
	in := `FOO=bar echo $FOO baz`
	p, err := ParseString(in)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.Commands) != 1 {
		t.Fatalf("expected 1 command, got %d", len(p.Commands))
	}
	cmd := p.Commands[0]
	if cmd.Name != "echo" {
		t.Fatalf("expected command name 'echo', got %q", cmd.Name)
	}
	if v := cmd.Assignments["FOO"]; v != "bar" {
		t.Fatalf("expected assignment FOO=bar, got %#v", cmd.Assignments)
	}
	if len(cmd.Args) != 2 || cmd.Args[0].Text != "$FOO" {
		t.Fatalf("unexpected args: %#v", cmd.Args)
	}
}

func TestEnvSubstitute(t *testing.T) {
	env := NewEnvFromOS()
	env.Set("X", "val")
	out := env.Substitute("$X and ${X} and $MISSING", 0)
	if out != "val and val and " {
		t.Fatalf("unexpected substitution result: %q", out)
	}
	// single quotes: no substitution
	out2 := env.Substitute("$X", '\'')
	if out2 != "$X" {
		t.Fatalf("single-quote should prevent substitution, got: %q", out2)
	}
}

func TestExecutePipelineEchoWc(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()
	out, code, err := runPipeline("echo \"a b\nc\" | wc", env)
	req.NoError(err, "execute pipeline error (out=%q)", out)
	req.Equal(0, code, "expected exit code 0")
	// Expecting 2 lines, 3 words, bytes depends on encoding; use contains check
	req.True(strings.HasPrefix(strings.TrimSpace(out), "2 "), "unexpected wc output: %q", out)
}

func TestAssignmentThenEcho(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()
	out, code, err := runPipeline("NAME=foo echo $NAME", env)
	req.NoError(err)
	req.Equal(0, code, "expected code 0")
	req.Equal("foo\n", out, "unexpected output")
	req.Equal("", env.Get("NAME"), "global env should not be changed by command-local assignment")
}

func TestCatFile(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()
	tmpDir := t.TempDir()
	fn := filepath.Join(tmpDir, "testfile.txt")
	content := "hello\nworld\n"
	err := os.WriteFile(fn, []byte(content), 0o644)
	req.NoError(err, "failed to write temp file")
	out, code, err := runPipeline("cat "+fn, env)
	req.NoError(err, "execute error (out=%q)", out)
	req.Equal(0, code, "expected code 0")
	req.Equal(content, out, "cat output mismatch")
}

func TestParsePipelineMultipleCommands(t *testing.T) {
	in := `echo a | wc | cat`
	p, err := ParseString(in)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(p.Commands) != 3 {
		t.Fatalf("expected 3 commands, got %d", len(p.Commands))
	}
	if p.Commands[0].Name != "echo" || p.Commands[1].Name != "wc" || p.Commands[2].Name != "cat" {
		t.Fatalf("unexpected pipeline commands: %#v", p.Commands)
	}
}

func TestSubstitutionInDoubleQuotes(t *testing.T) {
	env := NewEnvFromOS()
	env.Set("A", "X")
	out, code, err := runPipeline(`echo "value:$A"`, env)
	if err != nil {
		t.Fatalf("execute error: %v (out=%q)", err, out)
	}
	if code != 0 {
		t.Fatalf("expected code 0, got %d", code)
	}
	if strings.TrimSpace(out) != "value:X" {
		t.Fatalf("unexpected output: %q", out)
	}
}

// Integration: cat file through wc
func TestIntegration_CatPipeWc(t *testing.T) {
	env := NewEnvFromOS()
	tmp := t.TempDir()
	fn := filepath.Join(tmp, "f.txt")
	content := "one two\nthree\n"
	if err := os.WriteFile(fn, []byte(content), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	out, code, err := runPipeline("cat "+fn+" | wc", env)
	if err != nil {
		t.Fatalf("integration exec failed: %v (out=%q)", err, out)
	}
	if code != 0 {
		t.Fatalf("integration expected 0, got %d", code)
	}
	// should start with line count `2 `
	if !strings.HasPrefix(strings.TrimSpace(out), "2 ") {
		t.Fatalf("unexpected wc output: %q", out)
	}
}

// Integration: assignment-only command updates environment
func TestIntegration_AssignmentOnly(t *testing.T) {
	env := NewEnvFromOS()
	p, err := ParseString("FOO=bar")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	code, err := ExecutePipeline(p, env, nil, io.Discard, io.Discard)
	if err != nil && err != ErrExitRequested {
		t.Fatalf("execute: %v", err)
	}
	if code != 0 {
		t.Fatalf("expected 0 code for assignment-only, got %d", code)
	}
	if env.Get("FOO") != "bar" {
		t.Fatalf("expected FOO=bar in env, got %q", env.Get("FOO"))
	}
}

// Test that exit returns ErrExitRequested and passes code
func TestExitBuiltin(t *testing.T) {
	env := NewEnvFromOS()
	p, err := ParseString("exit 42")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var out bytes.Buffer
	code, err := ExecutePipeline(p, env, nil, &out, &out)
	if err != ErrExitRequested {
		t.Fatalf("expected ErrExitRequested, got %v", err)
	}
	if code != 42 {
		t.Fatalf("expected exit code 42, got %d", code)
	}
}

// Table-driven tests: various pipeline combinations and edge cases where the output
// of one command is fed into the next via pipes.
func TestPipelinesTableDriven(t *testing.T) {
	env := NewEnvFromOS()
	cases := []struct {
		name   string
		cmd    string
		want   string
		prefix bool // if true, check prefix of trimmed output
	}{
		{name: "echo_cat", cmd: `echo hello | cat`, want: "hello", prefix: false},
		{name: "echo_wc", cmd: `echo hello world | wc`, want: "1 2", prefix: true},
		{name: "echo_cat_wc", cmd: `echo hello | cat | wc`, want: "1 1", prefix: true},
		{name: "assign_and_pipe", cmd: `X=hi echo $X | wc`, want: "1", prefix: true},
		{name: "empty_input", cmd: ``, want: "", prefix: false},
		{name: "unclosed_quote", cmd: `echo "unclosed test`, want: "unclosed test", prefix: false},
		{name: "escaped_space", cmd: `echo foo\ bar | cat`, want: "foo bar", prefix: false},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			out, code, err := runPipeline(c.cmd, env)
			if err != nil {
				t.Fatalf("case %s: unexpected error: %v (out=%q)", c.name, err, out)
			}
			if code != 0 {
				t.Fatalf("case %s: expected code 0, got %d (out=%q)", c.name, code, out)
			}
			got := strings.TrimSpace(out)
			if c.prefix {
				if !strings.HasPrefix(got, c.want) {
					t.Fatalf("case %s: expected prefix %q, got %q", c.name, c.want, got)
				}
			} else {
				if got != c.want {
					t.Fatalf("case %s: expected %q, got %q", c.name, c.want, got)
				}
			}
		})
	}
}

// Table-driven parse-error/edge-case tests for pipelines and tokenization.
func TestPipelineParseErrors(t *testing.T) {
	cases := []struct {
		name string
		cmd  string
	}{
		{name: "leading_pipe", cmd: `| echo a`},
		{name: "trailing_pipe", cmd: `echo a |`},
		{name: "pipe_without_left", cmd: `| wc`},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			_, err := ParseString(c.cmd)
			if err == nil {
				t.Fatalf("case %s: expected parse error for input %q, got nil", c.name, c.cmd)
			}
		})
	}
}

// Larger integration table-driven tests combining files, pipes and assignments.
func TestIntegration_TableDriven(t *testing.T) {
	env := NewEnvFromOS()
	tmp := t.TempDir()
	fn := filepath.Join(tmp, "multi.txt")
	content := "alpha beta\ngamma delta epsilon\n"
	if err := os.WriteFile(fn, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	cases := []struct {
		name  string
		cmd   string
		check func(t *testing.T, out string, code int, err error)
	}{
		{
			name: "cat_to_wc",
			cmd:  "cat " + fn + " | wc",
			check: func(t *testing.T, out string, code int, err error) {
				if err != nil {
					t.Fatalf("unexpected err: %v (out=%q)", err, out)
				}
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				trim := strings.TrimSpace(out)
				if !strings.HasPrefix(trim, "2 ") {
					t.Fatalf("expected 2 lines prefix, got %q", trim)
				}
			},
		},
		{
			name: "assign_and_cat",
			cmd:  "MYVAR=val echo $MYVAR | cat",
			check: func(t *testing.T, out string, code int, err error) {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				if strings.TrimSpace(out) != "val" {
					t.Fatalf("expected 'val', got %q", out)
				}
			},
		},
		{
			name: "multiple_pipes",
			cmd:  "echo one two three | wc | cat",
			check: func(t *testing.T, out string, code int, err error) {
				if err != nil {
					t.Fatalf("unexpected err: %v", err)
				}
				if code != 0 {
					t.Fatalf("expected code 0, got %d", code)
				}
				// wc output should have words count 3
				if !strings.Contains(out, " 3 ") && !strings.Contains(out, " 3\n") {
					t.Fatalf("expected words count 3 in output, got %q", out)
				}
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			out, code, err := runPipeline(c.cmd, env)
			c.check(t, out, code, err)
		})
	}
}

// Tests for local assignment behavior: VAR=val cmd should be local to the command,
// assignment-only in a pipeline should be local to the pipeline, while a single
// assignment-only command (no pipeline) sets the global environment.
func TestLocalAssignmentsBehavior(t *testing.T) {
	req := require.New(t)

	t.Run("command-local-assignment", func(t *testing.T) {
		env := NewEnvFromOS()
		// ensure FOO is unset initially
		env.Set("FOO", "")
		out, code, err := runPipeline("FOO=bar echo $FOO", env)
		req.NoError(err, "execution failed (out=%q)", out)
		req.Equal(0, code)
		req.Equal("bar", strings.TrimSpace(out), "expected substituted value for command-local assignment")
		// global env should NOT be mutated by command-local assignment
		req.Equal("", env.Get("FOO"), "global env should not be changed by command-local assignment")
	})

	t.Run("assignment-only-in-pipeline_is_pipeline_local", func(t *testing.T) {
		env := NewEnvFromOS()
		env.Set("FOO", "")
		// Here the first element in pipeline is an assignment-only command,
		// it should affect the pipeline but not the global env.
		out, code, err := runPipeline("FOO=bar | echo $FOO", env)
		req.NoError(err, "execution failed (out=%q)", out)
		req.Equal(0, code)
		req.Equal("bar", strings.TrimSpace(out), "expected pipeline to see assignment-only value")
		// global env should still be untouched
		req.Equal("", env.Get("FOO"), "global env should not be changed by pipeline-local assignment-only")
	})

	t.Run("single-assignment_sets_global", func(t *testing.T) {
		env := NewEnvFromOS()
		env.Set("FOO", "")
		// Single assignment-only command (no pipeline) should set the global env
		// According to earlier design, a single assignment command applies globally.
		p, err := ParseString("FOO=bar")
		req.NoError(err)
		code, err := ExecutePipeline(p, env, nil, io.Discard, io.Discard)
		req.NoError(err)
		req.Equal(0, code)
		req.Equal("bar", env.Get("FOO"), "single assignment-only command should set global env")
	})
}

// Additional substitution-focused table-driven tests.
func TestSubstitutionCombinations(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()
	env.Set("A", "1")
	env.Set("B", "two")
	env.Set("EMPTY", "")

	cases := []struct {
		name string
		cmd  string
		want string
	}{
		{
			name: "simple_var",
			cmd:  `echo $A`,
			want: "1",
		},
		{
			name: "braced_var",
			cmd:  `echo ${B}`,
			want: "two",
		},
		{
			name: "concat_var",
			cmd:  `echo pre$Apost`,
			want: "pre", // $Apost is not defined -> empty, so output "pre"
		},
		{
			name: "multiple_vars",
			cmd:  `echo $A-$B`,
			want: "1-two",
		},
		{
			name: "empty_var",
			cmd:  `echo $EMPTY`,
			want: "",
		},
		{
			name: "single_quotes_no_sub",
			cmd:  `echo '$A $B'`,
			want: "$A $B",
		},
		{
			name: "double_quotes_with_sub",
			cmd:  `echo "$A $B"`,
			want: "1 two",
		},
		{
			name: "missing_var_results_empty",
			cmd:  `echo $NO_SUCH`,
			want: "",
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			out, code, err := runPipeline(c.cmd, env)
			req.NoError(err, "case %s failed with error", c.name)
			req.Equal(0, code, "case %s expected code 0 got %d", c.name, code)
			got := strings.TrimSpace(out)
			req.Equal(c.want, got, "case %s: expected %q got %q", c.name, c.want, got)
		})
	}
}

// Table-driven tests ensuring data flow through pipelines: output of one is input to next.
func TestPipelineDataFlow(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	f := filepath.Join(tmp, "flow.txt")
	err := os.WriteFile(f, []byte("alpha\nbeta gamma\n"), 0644)
	req.NoError(err)

	cases := []struct {
		name  string
		cmd   string
		check func(t *testing.T, out string, code int, err error)
	}{
		{
			name: "echo_to_wc_counts",
			cmd:  `echo "one two three" | wc`,
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				trim := strings.TrimSpace(out)
				// expect 1 line and 3 words as prefix
				req.True(strings.HasPrefix(trim, "1 3 "), "unexpected wc output: %q", trim)
			},
		},
		{
			name: "file_cat_pipe_wc",
			cmd:  `cat ` + f + ` | wc`,
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				trim := strings.TrimSpace(out)
				// 2 lines in file
				req.True(strings.HasPrefix(trim, "2 "), "unexpected wc output: %q", trim)
			},
		},
		{
			name: "echo_pipe_cat_dash_wc",
			cmd:  `echo "hello world" | cat - | wc`,
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				trim := strings.TrimSpace(out)
				req.True(strings.HasPrefix(trim, "1 2 "), "unexpected wc output: %q", trim)
			},
		},
		{
			// chaining transforms: echo -> awk-like emulation via external command not available;
			// use builtin transformations: echo -> cat -> wc
			name: "multiple_stage_pipeline",
			cmd:  `echo "a b" | cat | cat | wc`,
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				trim := strings.TrimSpace(out)
				req.True(strings.HasPrefix(trim, "1 2 "), "unexpected wc output: %q", trim)
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			out, code, err := runPipeline(c.cmd, env)
			c.check(t, out, code, err)
		})
	}
}

// Additional table-driven tests covering more combinations and edge cases.
func TestMorePipelines_TableDriven(t *testing.T) {

	env := NewEnvFromOS()
	req := require.New(t)

	// prepare files for multi-file wc test
	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.txt")
	f2 := filepath.Join(tmp, "b.txt")
	_ = os.WriteFile(f1, []byte("one two\n"), 0644)
	_ = os.WriteFile(f2, []byte("three four five\n"), 0644)

	cases := []struct {
		name  string
		cmd   string
		check func(t *testing.T, out string, code int, err error)
	}{
		{
			name: "wc_multiple_files",
			cmd:  "wc " + f1 + " " + f2,
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				// must contain both filenames and a total line
				req.Contains(out, "a.txt")
				req.Contains(out, "b.txt")
				req.Contains(out, "total")
			},
		},
		{
			name: "exit_in_pipeline",
			cmd:  "echo hi | exit 7 | cat",
			check: func(t *testing.T, out string, code int, err error) {
				// when exit happens inside pipeline we expect ErrExitRequested
				req.ErrorIs(err, ErrExitRequested)
				req.Equal(7, code)
			},
		},
		{
			name: "assignments_before_command",
			cmd:  "A=1 B=2 echo $A $B",
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				req.Equal("1 2", strings.TrimSpace(out))
			},
		},
		{
			name: "cat_dash_reads_stdin",
			cmd:  "echo hello | cat -",
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				req.Equal("hello", strings.TrimSpace(out))
			},
		},
		{
			name: "single_quote_no_sub",
			cmd:  "X=val echo '$X'",
			check: func(t *testing.T, out string, code int, err error) {
				req.NoError(err)
				req.Equal(0, code)
				req.Equal("$X", strings.TrimSpace(out))
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			out, code, err := runPipeline(c.cmd, env)
			c.check(t, out, code, err)
		})
	}
}
