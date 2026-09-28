package checks

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ExecResult is the outcome of running a command.
type ExecResult struct {
	// Output is the tail of combined stdout and stderr.
	Output   []byte
	ExitCode int
	// NotFound is set when the program is not installed.
	NotFound bool
	// TimedOut is set when the command exceeded its deadline.
	TimedOut bool
	// Err is set when the command could not be started or waited on.
	Err error
}

// Executor runs a command in a directory. Tests replace it with a fake.
type Executor interface {
	Run(ctx context.Context, dir string, args []string, env []string) ExecResult
}

// SystemExecutor runs real processes. Commands are started directly,
// never through a shell, with stdin closed.
type SystemExecutor struct{}

// maxOutput is how much trailing output is kept per command.
const maxOutput = 64 << 10

// Run implements Executor.
func (SystemExecutor) Run(ctx context.Context, dir string, args []string, env []string) ExecResult {
	if len(args) == 0 {
		return ExecResult{Err: errors.New("empty command")}
	}
	program := args[0]
	// Project-relative programs such as ./gradlew resolve against dir.
	if strings.ContainsAny(program, `/\`) && !filepath.IsAbs(program) {
		program = filepath.Join(dir, program)
	}
	path, err := exec.LookPath(program)
	if err != nil {
		return ExecResult{NotFound: true, ExitCode: -1, Err: err}
	}

	cmd := exec.Command(path, args[1:]...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdin = nil
	out := &tailBuffer{max: maxOutput}
	cmd.Stdout = out
	cmd.Stderr = out
	// Don't hang if the command leaves a child process holding the output
	// pipe open after it exits.
	cmd.WaitDelay = 5 * time.Second
	setProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return ExecResult{ExitCode: -1, Err: err}
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	var waitErr error
	timedOut := false
	select {
	case waitErr = <-done:
	case <-ctx.Done():
		timedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
		killProcessGroup(cmd)
		select {
		case waitErr = <-done:
		case <-time.After(5 * time.Second):
			waitErr = ctx.Err()
		}
	}

	res := ExecResult{Output: out.Bytes(), TimedOut: timedOut}
	if ctx.Err() != nil && !timedOut {
		res.Err = ctx.Err()
		res.ExitCode = -1
		return res
	}
	var exitErr *exec.ExitError
	switch {
	case waitErr == nil:
	case errors.Is(waitErr, exec.ErrWaitDelay):
		res.ExitCode = cmd.ProcessState.ExitCode()
	case errors.As(waitErr, &exitErr):
		res.ExitCode = exitErr.ExitCode()
	default:
		res.ExitCode = -1
		if !timedOut {
			res.Err = waitErr
		}
	}
	return res
}

// tailBuffer keeps only the last max bytes written to it.
type tailBuffer struct {
	mu  sync.Mutex
	buf []byte
	max int
}

func (t *tailBuffer) Write(p []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, p...)
	if len(t.buf) > t.max {
		t.buf = append([]byte(nil), t.buf[len(t.buf)-t.max:]...)
	}
	return len(p), nil
}

func (t *tailBuffer) Bytes() []byte {
	t.mu.Lock()
	defer t.mu.Unlock()
	return bytes.Clone(t.buf)
}
