//go:build linux

package container

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sevladev/minic/internal/namespace"
)

const imageBaseDir = "/var/lib/minic/images"

func Run(cfg Config) error {
	rootfs, err := resolveRootfs(cfg.Image)
	if err != nil {
		return err
	}

	args := append([]string{"init"}, cfg.Command...)
	cmd := exec.Command("/proc/self/exe", args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = namespace.NewSysProcAttr()

	hostname := cfg.Hostname
	if hostname == "" {
		hostname = "minic"
	}

	cmd.Env = append(os.Environ(),
		"MINIC_HOSTNAME="+hostname,
		"MINIC_ROOTFS="+rootfs,
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("container exited: %w", err)
	}

	return nil
}

func resolveRootfs(image string) (string, error) {
	rootfs := filepath.Join(imageBaseDir, image, "rootfs")

	if _, err := os.Stat(rootfs); os.IsNotExist(err) {
		return "", fmt.Errorf("image %q not found at %s\nRun: minic pull %s", image, rootfs, image)
	}

	return rootfs, nil
}
