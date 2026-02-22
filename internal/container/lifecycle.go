//go:build linux

package container

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/sevladev/minic/internal/cgroup"
	"github.com/sevladev/minic/internal/namespace"
)

const imageBaseDir = "/var/lib/minic/images"

func Run(cfg Config) error {
	rootfs, err := resolveRootfs(cfg.Image)
	if err != nil {
		return err
	}

	containerID := generateID()

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

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	hasCgroup := false
	limits := cgroup.Limits{
		MemoryBytes: cfg.Resources.MemoryBytes,
		CPUQuota:    cfg.Resources.CPUQuota,
		PidsMax:     cfg.Resources.PidsMax,
	}
	if limits.MemoryBytes > 0 || limits.CPUQuota > 0 || limits.PidsMax > 0 {
		if err := cgroup.Apply(containerID, cmd.Process.Pid, limits); err != nil {
			cmd.Process.Kill()
			return fmt.Errorf("apply cgroups: %w", err)
		}
		hasCgroup = true
	}

	err = cmd.Wait()

	if hasCgroup {
		cgroup.Remove(containerID)
	}

	if err != nil {
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

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
