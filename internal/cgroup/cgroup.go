//go:build linux

package cgroup

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const basePath = "/sys/fs/cgroup/minic"

type Limits struct {
	MemoryBytes int64
	CPUQuota    float64
	PidsMax     int
}

func cgroupPath(containerID string) string {
	return filepath.Join(basePath, "container-"+containerID)
}

func Apply(containerID string, pid int, limits Limits) error {
	if err := enableControllers("/sys/fs/cgroup"); err != nil {
		return fmt.Errorf("enable controllers on root: %w", err)
	}

	if err := os.MkdirAll(basePath, 0755); err != nil {
		return fmt.Errorf("create minic cgroup: %w", err)
	}

	if err := enableControllers(basePath); err != nil {
		return fmt.Errorf("enable controllers on minic: %w", err)
	}

	path := cgroupPath(containerID)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("create container cgroup: %w", err)
	}

	if limits.MemoryBytes > 0 {
		if err := writeFile(filepath.Join(path, "memory.max"), strconv.FormatInt(limits.MemoryBytes, 10)); err != nil {
			return fmt.Errorf("set memory.max: %w", err)
		}
	}

	if limits.CPUQuota > 0 {
		quota := int(limits.CPUQuota * 100000)
		val := fmt.Sprintf("%d 100000", quota)
		if err := writeFile(filepath.Join(path, "cpu.max"), val); err != nil {
			return fmt.Errorf("set cpu.max: %w", err)
		}
	}

	if limits.PidsMax > 0 {
		if err := writeFile(filepath.Join(path, "pids.max"), strconv.Itoa(limits.PidsMax)); err != nil {
			return fmt.Errorf("set pids.max: %w", err)
		}
	}

	if err := writeFile(filepath.Join(path, "cgroup.procs"), strconv.Itoa(pid)); err != nil {
		return fmt.Errorf("add pid to cgroup: %w", err)
	}

	return nil
}

func enableControllers(path string) error {
	return writeFile(filepath.Join(path, "cgroup.subtree_control"), "+memory +cpu +pids")
}

func Remove(containerID string) error {
	return os.Remove(cgroupPath(containerID))
}

func writeFile(path, value string) error {
	return os.WriteFile(path, []byte(value), 0644)
}
