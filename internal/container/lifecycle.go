//go:build linux

package container

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/sevladev/minic/internal/namespace"
)

func Run(cfg Config) error {
	args := append([]string{"init"}, cfg.Command...)
	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = namespace.NewSysProcAttr()

	if cfg.Hostname != "" {
		cmd.Env = append(os.Environ(), "MINIC_HOSTNAME="+cfg.Hostname)
	} else {
		cmd.Env = append(os.Environ(), "MINIC_HOSTNAME=minic")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container exited: %w", err)
	}

	return nil
}
