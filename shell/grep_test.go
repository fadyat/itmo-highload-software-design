package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Tests for the builtin `grep` implementation added to the shell.
// These tests exercise:
//  - basic regex matching
//  - -w (whole-word) flag behavior
//  - -i (case-insensitive) flag
//  - -A (after) flag and overlapping contexts
//  - reading from stdin when no files are provided
//  - prefixing of output when multiple files are searched
//  - invalid regexp handling
//
// Note: tests use the test helper `runPipeline` defined in exec_test.go
// which parses a command string, executes the pipeline and captures stdout/stderr.

func TestGrep_WordFlag(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	fn := filepath.Join(tmp, "word.txt")
	content := strings.Join([]string{
		"Minimal",         // should match
		"Minimally",       // should NOT match when -w
		"Another Minimal", // should match
		"foo_Minimal_bar", // underscores are word chars -> should NOT match with -w
		"Minimal.",        // punctuation boundary -> should match
	}, "\n") + "\n"
	err := os.WriteFile(fn, []byte(content), 0o644)
	req.NoError(err)

	out, code, err := runPipeline("grep -w Minimal "+fn, env)
	req.NoError(err)
	req.Equal(0, code)

	lines := strings.Split(strings.TrimSpace(out), "\n")
	// Expect lines: Minimal, Another Minimal, Minimal.
	// Order preserved; there are 3 matching logical lines.
	req.Equal(3, len(lines))
	req.Contains(lines[0], "Minimal")
	req.Contains(lines[1], "Another Minimal")
	req.Contains(lines[2], "Minimal")
}

func TestGrep_CaseInsensitive(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	fn := filepath.Join(tmp, "case.txt")
	content := strings.Join([]string{
		"MINIMAL",
		"minimal",
		"MiNiMaL in the middle",
		"no match here",
	}, "\n") + "\n"
	err := os.WriteFile(fn, []byte(content), 0o644)
	req.NoError(err)

	out, code, err := runPipeline("grep -i minimal "+fn, env)
	req.NoError(err)
	req.Equal(0, code)

	got := strings.TrimSpace(out)
	lines := strings.Split(got, "\n")
	// Expect three matching lines
	req.Equal(3, len(lines))
}

func TestGrep_AfterFlagAndOverlap(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	fn := filepath.Join(tmp, "after.txt")
	content := strings.Join([]string{
		"line1",
		"match", // 2
		"line3", // context of match@2
		"match", // 4
		"line5", // context of match@4
		"line6",
	}, "\n") + "\n"
	err := os.WriteFile(fn, []byte(content), 0o644)
	req.NoError(err)

	out, code, err := runPipeline("grep -A 1 match "+fn, env)
	req.NoError(err)
	req.Equal(0, code)

	// Expect printed lines: match (line2), line3, match (line4), line5
	got := strings.TrimSpace(out)
	lines := strings.Split(got, "\n")
	req.Equal(4, len(lines))
	req.Equal("match", strings.TrimSpace(lines[0]))
	req.Equal("line3", strings.TrimSpace(lines[1]))
	req.Equal("match", strings.TrimSpace(lines[2]))
	req.Equal("line5", strings.TrimSpace(lines[3]))

	// Now test overlapping matches (matches on consecutive lines)
	fn2 := filepath.Join(tmp, "overlap.txt")
	content2 := strings.Join([]string{
		"m1",
		"match", // 2
		"match", // 3 (consecutive)
		"ctx",   // context of match@3 or match@2+A1
	}, "\n") + "\n"
	err = os.WriteFile(fn2, []byte(content2), 0o644)
	req.NoError(err)

	out2, code2, err2 := runPipeline("grep -A 1 match "+fn2, env)
	req.NoError(err2)
	req.Equal(0, code2)

	lines2 := strings.Split(strings.TrimSpace(out2), "\n")
	// Should print: match (2), match (3), ctx (4) -- no duplicates
	req.Equal(3, len(lines2))
	req.Equal("match", strings.TrimSpace(lines2[0]))
	req.Equal("match", strings.TrimSpace(lines2[1]))
	req.Equal("ctx", strings.TrimSpace(lines2[2]))
}

func TestGrep_ReadFromStdin(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// echo a multi-line string into grep via a pipeline
	out, code, err := runPipeline(`echo "a
match
b" | grep match`, env)
	req.NoError(err)
	req.Equal(0, code)
	req.Equal("match", strings.TrimSpace(out))
}

func TestGrep_MultipleFilesPrefixing(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	f1 := filepath.Join(tmp, "a.txt")
	f2 := filepath.Join(tmp, "b.txt")
	_ = os.WriteFile(f1, []byte("one\npattern here\n"), 0o644)
	_ = os.WriteFile(f2, []byte("pattern in second\nother\n"), 0o644)

	out, code, err := runPipeline("grep pattern "+f1+" "+f2, env)
	req.NoError(err)
	req.Equal(0, code)

	trim := strings.TrimSpace(out)
	req.Contains(trim, "a.txt:pattern here")
	req.Contains(trim, "b.txt:pattern in second")
}

func TestGrep_InvalidRegexpProducesError(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	fn := filepath.Join(tmp, "dummy.txt")
	_ = os.WriteFile(fn, []byte("line\n"), 0o644)

	// Use an invalid regexp: unclosed character class
	out, code, err := runPipeline("grep '[' "+fn, env)
	// Expect an error (non-nil) and non-zero code
	req.Error(err)
	// code may be non-zero (implementation returns 2 on invalid regexp)
	req.NotEqual(0, code)
	// Output should mention invalid regexp or similar message
	req.True(strings.Contains(out, "invalid regexp") || strings.Contains(out, "invalid"), "expected error message in output, got: %q", out)
}

// New tests combining grep with pipelines and other builtins (echo, cat, wc, and '-' stdin handling).

func TestGrep_PipeEchoCat(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// echo -> cat -> grep pipeline
	out, code, err := runPipeline(`echo "aa
pattern
bb" | cat | grep pattern`, env)
	req.NoError(err)
	req.Equal(0, code)
	req.Equal("pattern", strings.TrimSpace(out))
}

func TestGrep_CatDashReadsStdin(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	tmp := t.TempDir()
	fn := filepath.Join(tmp, "inp.txt")
	_ = os.WriteFile(fn, []byte("one\npattern here\ntwo\n"), 0o644)

	// Use '-' to tell grep to read from stdin; cat provides stdin
	out, code, err := runPipeline("cat "+fn+" | grep pattern -", env)
	req.NoError(err)
	req.Equal(0, code)
	req.Equal("pattern here", strings.TrimSpace(out))
}

func TestGrep_PipeGrepToWc(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// produce multiline input, filter with grep, then wc the result
	out, code, err := runPipeline(`echo "one\npattern\nalpha" | grep pattern | wc`, env)
	req.NoError(err)
	req.Equal(0, code)
	trim := strings.TrimSpace(out)
	// grep should have emitted one line with one word -> wc output should start with "1 1 "
	req.True(strings.HasPrefix(trim, "1 1 "), "unexpected wc output: %q", trim)
}
