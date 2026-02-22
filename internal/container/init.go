//go:build linux

package container

import (
	"fmt"
	"os"
	"syscall"
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

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc: %w", err)
	}

	binary, err := exec_lookpath(args[0])
	if err != nil {
		return fmt.Errorf("exec: %w", err)
	}

	return syscall.Exec(binary, args, os.Environ())
}

func exec_lookpath(file string) (string, error) {
	if file[0] == '/' {
		return file, nil
	}

	paths := []string{"/usr/local/sbin", "/usr/local/bin", "/usr/sbin", "/usr/bin", "/sbin", "/bin"}
	for _, dir := range paths {
		path := dir + "/" + file
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return file, nil
}
