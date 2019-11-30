package proclimit

import (
	"bytes"
	"github.com/friendsofgo/errors"
	"os/exec"
	"testing"
)

func echo(args ...string) *exec.Cmd {
	return exec.Command("echo", args...)
}

type spyLimiter struct {
	calledWithPid int
	returnErr     error
}

func (s *spyLimiter) Limit(pid int) error {
	s.calledWithPid = pid
	return s.returnErr
}

func TestCmdStartInvokesLimit(t *testing.T) {
	sl := &spyLimiter{}
	cmd := &Cmd{Cmd: echo(), Limiter: sl}
	err := cmd.Start()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if sl.calledWithPid != cmd.Process.Pid {
		t.Errorf("expected pid %d to be limited, but got %d", cmd.Process.Pid, sl.calledWithPid)
	}
}

func TestCmdStartLimiterErrors(t *testing.T) {
	sl := &spyLimiter{returnErr: errors.New("limit error")}
	cmd := &Cmd{Cmd: echo(), Limiter: sl}
	err := cmd.Start()
	if err == nil || errors.Cause(err) != sl.returnErr {
		t.Errorf("expected error \"%v\", but got: %v", sl.returnErr, err)
	}
	s, err := cmd.Process.Wait()
	if err != nil {
		t.Fatalf("failed to wait for process: %v", err)
	}
	if s.ExitCode() != -1 {
		t.Errorf("expected exit code -1 (terminated by signal), but got: %d", s.ExitCode())
	}
}

func TestCmdOutput(t *testing.T) {
	cmd := &Cmd{Cmd: echo("hello, world!"), Limiter: &spyLimiter{}}
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("expected no error, but got: %v", err)
	}
	if !bytes.Equal(out, []byte("hello, world!\n")) {
		t.Errorf("expected 'hello, world!\\n', but got: '%s'", string(out))
	}
}
