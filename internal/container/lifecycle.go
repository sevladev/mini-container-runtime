//go:build linux

package container

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"

	"github.com/sevladev/minic/internal/cgroup"
	"github.com/sevladev/minic/internal/filesystem"
	"github.com/sevladev/minic/internal/image"
	"github.com/sevladev/minic/internal/namespace"
	"github.com/sevladev/minic/internal/network"
)

func Run(cfg Config) error {
	if !image.Exists(cfg.Image) {
		return fmt.Errorf("image %q not found\nRun: minic pull %s", cfg.Image, cfg.Image)
	}

	containerID := generateID()

	overlay, err := filesystem.SetupOverlay(containerID, image.RootfsPath(cfg.Image))
	if err != nil {
		return fmt.Errorf("setup overlay: %w", err)
	}
	defer filesystem.RemoveOverlay(overlay)

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
		"MINIC_ROOTFS="+overlay.Merged,
	)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start container: %w", err)
	}

	pid := cmd.Process.Pid

	if cfg.NetMode != "none" {
		net, err := network.Setup(containerID, pid)
		if err != nil {
			cmd.Process.Kill()
			return fmt.Errorf("setup network: %w", err)
		}
		defer network.Cleanup(net.ContainerID)
		fmt.Printf("Container IP: %s\n", net.IP)
	}

	hasCgroup := false
	limits := cgroup.Limits{
		MemoryBytes: cfg.Resources.MemoryBytes,
		CPUQuota:    cfg.Resources.CPUQuota,
		PidsMax:     cfg.Resources.PidsMax,
	}
	if limits.MemoryBytes > 0 || limits.CPUQuota > 0 || limits.PidsMax > 0 {
		if err := cgroup.Apply(containerID, pid, limits); err != nil {
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

func generateID() string {
	b := make([]byte, 6)
	rand.Read(b)
	return hex.EncodeToString(b)
}
