//go:build linux

package namespace

import (
	"os"
	"os/exec"
	"strconv"
)

func EnterAndExec(pid int, command []string) error {
	pidStr := strconv.Itoa(pid)

	args := []string{
		"-t", pidStr,
		"-m", "-u", "-i", "-n", "-p",
		"--",
	}
	args = append(args, command...)

	cmd := exec.Command("nsenter", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
