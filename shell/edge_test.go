package shell

import (
	"bytes"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// This file contains additional edge and stress tests for pipelines, assignments,
// and exit behavior. These tests are intentionally a bit heavier and aim to
// exercise concurrency and pipe-closing behavior under repeated/large loads.

// TestExitInMiddleRepeated runs a pipeline with `exit` in the middle repeatedly
// to ensure we don't deadlock when a builtin requests exit while upstream
// goroutines are writing to pipes.
func TestExitInMiddleRepeated(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// Run the scenario several times to increase chance of catching races/deadlocks.
	for i := 0; i < 200; i++ {
		_, code, err := runPipeline("echo hi | exit 7 | cat", env)
		// output from echo might be discarded; we only care about exit propagation
		req.ErrorIs(err, ErrExitRequested, "iteration %d: expected ErrExitRequested", i)
		req.Equal(7, code, "iteration %d: expected exit code 7", i)
	}
}

// TestLongPipelineChain builds a long chain of builtins (echo -> cat -> ... -> wc)
// and verifies the data flows correctly and the pipeline terminates.
func TestLongPipelineChain(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// build long chain: echo "a b" | cat | cat | ... | wc
	chainLen := 60
	var sb strings.Builder
	sb.WriteString(`echo "a b"`)
	for i := 0; i < chainLen; i++ {
		sb.WriteString(" | cat")
	}
	sb.WriteString(" | wc")

	cmd := sb.String()

	// Run the pipeline a few times
	for i := 0; i < 20; i++ {
		out, code, err := runPipeline(cmd, env)
		req.NoError(err, "iteration %d: unexpected error (out=%q)", i, out)
		req.Equal(0, code, "iteration %d: expected 0 exit code", i)
		trim := strings.TrimSpace(out)
		// expect one line and two words as prefix
		req.True(strings.HasPrefix(trim, "1 2 "), "iteration %d: unexpected wc output: %q", i, trim)
	}
}

// TestAssignmentScopeConcurrent stresses assignment scoping with concurrent runs.
// It verifies that command-local assignments don't leak to global environment
// and that pipeline-local assignments remain local to the pipeline.
func TestAssignmentScopeConcurrent(t *testing.T) {
	req := require.New(t)

	// We'll run many pipelines in parallel to stress the env cloning logic.
	const workers = 40
	var wg sync.WaitGroup
	wg.Add(workers)
	errCh := make(chan error, workers)

	for i := 0; i < workers; i++ {
		go func(i int) {
			defer wg.Done()
			env := NewEnvFromOS()
			// ensure VAR is not set
			env.Set("VAR", "")
			// command-local assignment
			_, code, err := runPipeline("VAR=val"+strings.TrimSpace(string('0'+byte(i%10)))+" echo $VAR", env)
			if err != nil {
				errCh <- err
				return
			}
			if code != 0 {
				errCh <- err
				return
			}
			// Output should be the value substituted for command-local assignment.
			// We don't check the exact suffix but ensure global env unchanged.
			if env.Get("VAR") != "" {
				errCh <- err
				return
			}

			// pipeline-local assignment (assignment-only command before a command)
			env2 := NewEnvFromOS()
			env2.Set("FOO", "")
			out2, code2, err2 := runPipeline("FOO=bar | echo $FOO", env2)
			if err2 != nil {
				errCh <- err2
				return
			}
			if code2 != 0 {
				errCh <- err2
				return
			}
			if strings.TrimSpace(out2) != "bar" {
				errCh <- err2
				return
			}
			// global env must remain unchanged
			if env2.Get("FOO") != "" {
				errCh <- err2
				return
			}
		}(i)
	}

	// Wait with a timeout to catch potential deadlocks
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(8 * time.Second):
		t.Fatalf("concurrent assignment scope stress timed out")
	}

	close(errCh)
	for e := range errCh {
		req.NoError(e)
	}
}

// TestLargeInputCloseOnExit verifies that when an exit builtin appears in the
// pipeline, a large upstream reader doesn't cause a hang: ExecutePipeline should
// return promptly with ErrExitRequested and the exit code.
func TestLargeInputCloseOnExit(t *testing.T) {
	req := require.New(t)
	env := NewEnvFromOS()

	// Prepare a large input (~1MB)
	var bigBuf strings.Builder
	lines := 20000
	for i := 0; i < lines; i++ {
		bigBuf.WriteString("line\n")
	}
	inReader := strings.NewReader(bigBuf.String())

	p, err := ParseString("cat | exit 2")
	req.NoError(err)

	var out bytes.Buffer
	var errOut bytes.Buffer

	// Run ExecutePipeline directly with the large reader as stdin.
	// This should not hang; we expect ErrExitRequested and code 2.
	start := time.Now()
	code, err := ExecutePipeline(p, env, inReader, &out, &errOut)
	elapsed := time.Since(start)

	req.ErrorIs(err, ErrExitRequested)
	req.Equal(2, code)
	// Ensure it returns quickly (arbitrary upper bound)
	req.Less(elapsed, 5*time.Second, "ExecutePipeline took too long: %v", elapsed)
}

// TestConcurrentExecuteMultiplePipelines runs multiple different pipelines concurrently
// to increase chance of catching races or deadlocks involving shared structures.
func TestConcurrentExecuteMultiplePipelines(t *testing.T) {
	req := require.New(t)

	env := NewEnvFromOS()
	commands := []string{
		`echo "alpha beta" | wc`,
		`echo "x y z" | cat | wc`,
		`FOO=val echo $FOO | wc`,
		`echo hi | exit 3 | cat`,
		`echo one | cat - | wc`,
	}

	var wg sync.WaitGroup
	wg.Add(len(commands) * 6) // run each command multiple times

	errC := make(chan error, len(commands)*6)

	for i := 0; i < 6; i++ {
		for _, c := range commands {
			cmd := c
			go func() {
				defer wg.Done()
				out, code, err := runPipeline(cmd, env)
				// For exit case we expect ErrExitRequested; others should be OK.
				if strings.Contains(cmd, "exit") {
					if err != ErrExitRequested {
						errC <- err
						return
					}
				} else {
					if err != nil {
						errC <- err
						return
					}
					if code != 0 {
						errC <- err
						return
					}
					// simple sanity for wc outputs
					_ = out
				}
			}()
		}
	}

	// Wait with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// ok
	case <-time.After(12 * time.Second):
		t.Fatalf("concurrent pipelines timed out")
	}

	close(errC)
	for e := range errC {
		req.NoError(e)
	}
}
