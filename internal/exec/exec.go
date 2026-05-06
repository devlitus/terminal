package exec

import (
	"bufio"
	"context"
	"errors"
	"os"
	osExec "os/exec"
	"runtime"
	"sync"
	"time"
)

// Run executes command in dir, streaming output lines to out.
// Returns exitCode, wall-clock duration, and any non-exit error.
// The out channel is closed exactly once when the command terminates.
func Run(ctx context.Context, dir string, command string, out chan<- string) (exitCode int, duration time.Duration, err error) {
	shell, args := resolveShell(command)
	cmd := osExec.CommandContext(ctx, shell, args...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 0, 0, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 0, 0, err
	}

	start := time.Now()
	if err = cmd.Start(); err != nil {
		return 0, 0, err
	}

	var wg sync.WaitGroup
	wg.Add(2)

	scan := func(r interface{ Read([]byte) (int, error) }) {
		defer wg.Done()
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}

	go scan(stdout)
	go scan(stderr)

	waitErr := cmd.Wait()
	duration = time.Since(start)
	wg.Wait()
	close(out)

	var exitErr *osExec.ExitError
	if errors.As(waitErr, &exitErr) {
		return exitErr.ExitCode(), duration, nil
	}
	return 0, duration, waitErr
}

func resolveShell(command string) (shell string, args []string) {
	if runtime.GOOS == "windows" {
		return "cmd.exe", []string{"/C", command}
	}
	shell = os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return shell, []string{"-c", command}
}
