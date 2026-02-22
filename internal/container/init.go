//go:build linux

package container

import (
	"fmt"
	"os"
	"syscall"

	"github.com/sevladev/minic/internal/filesystem"
)

func Init(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("init: no command provided")
	}

	hostname := os.Getenv("MINIC_HOSTNAME")
	if hostname != "" {
		if err := syscall.Sethostname([]byte(hostname)); err != nil {
			return fmt.Errorf("sethostname: %w", err)
		}
	}

	rootfs := os.Getenv("MINIC_ROOTFS")
	if rootfs != "" {
		if err := syscall.Mount("", "/", "", syscall.MS_REC|syscall.MS_PRIVATE, ""); err != nil {
			return fmt.Errorf("make root private: %w", err)
		}
		if err := filesystem.PivotRoot(rootfs); err != nil {
			return fmt.Errorf("pivot_root: %w", err)
		}
		if err := filesystem.SetupMounts(); err != nil {
			return fmt.Errorf("setup mounts: %w", err)
		}
	}

	binary := lookpath(args[0])
	return syscall.Exec(binary, args, os.Environ())
}

func lookpath(file string) string {
	if len(file) > 0 && file[0] == '/' {
		return file
	}

	dirs := []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}
	for _, dir := range dirs {
		path := dir + "/" + file
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return file
}
