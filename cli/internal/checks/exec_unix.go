//go:build !windows

package checks

import (
	"os/exec"
	"syscall"
)

// setProcessGroup starts the command in its own process group so that a
// timeout also stops any children it spawned (npm → node, gradle → java).
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
