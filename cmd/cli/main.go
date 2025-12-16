package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/fadyat/cli/shell"
)

func main() {
	env := shell.NewEnvFromOS()
	reader := bufio.NewReader(os.Stdin)

	for {
		// Print prompt
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF or other read error — exit gracefully
			if err == os.ErrClosed || err == io.EOF {
				fmt.Println()
				return
			}
			fmt.Fprintln(os.Stderr, "read error:", err)
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Quick builtin: allow user to type "quit" to exit interactively.
		// The executor also supports `exit` builtin inside pipelines/commands.
		if line == "quit" || line == "q" {
			return
		}

		// Parse input into pipeline
		pipeline, err := shell.ParseString(line)
		if err != nil {
			fmt.Fprintln(os.Stderr, "parse error:", err)
			continue
		}

		// Execute pipeline. We pass os.Stdout/os.Stderr so command output goes to terminal.
		code, err := shell.ExecutePipeline(pipeline, env, nil, os.Stdout, os.Stderr)
		if err != nil {
			// Special handling for builtin exit
			if err == shell.ErrExitRequested {
				os.Exit(code)
			}
			fmt.Fprintln(os.Stderr, "execution error:", err)
			// continue loop even on non-fatal errors
			continue
		}

		// Non-zero exit code from last command: print status (optional)
		if code != 0 {
			fmt.Fprintf(os.Stderr, "exit code: %d\n", code)
		}
	}
}
