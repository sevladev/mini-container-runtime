//go:build linux

package filesystem

import (
	"fmt"
	"os"
	"syscall"
)

func SetupMounts() error {
	if err := mountProc(); err != nil {
		return err
	}
	if err := mountSys(); err != nil {
		return err
	}
	if err := mountDev(); err != nil {
		return err
	}
	if err := createDevices(); err != nil {
		return err
	}
	if err := mountDevPts(); err != nil {
		return err
	}
	return setupResolv()
}

func mountProc() error {
	os.MkdirAll("/proc", 0755)
	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		return fmt.Errorf("mount /proc: %w", err)
	}
	return nil
}

func mountSys() error {
	os.MkdirAll("/sys", 0755)
	if err := syscall.Mount("sysfs", "/sys", "sysfs", syscall.MS_RDONLY, ""); err != nil {
		return fmt.Errorf("mount /sys: %w", err)
	}
	return nil
}

func mountDev() error {
	os.MkdirAll("/dev", 0755)
	if err := syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755"); err != nil {
		return fmt.Errorf("mount /dev: %w", err)
	}
	return nil
}

type device struct {
	path  string
	major uint32
	minor uint32
	mode  uint32
}

func createDevices() error {
	devices := []device{
		{"/dev/null", 1, 3, 0666},
		{"/dev/zero", 1, 5, 0666},
		{"/dev/random", 1, 8, 0444},
		{"/dev/urandom", 1, 9, 0444},
		{"/dev/tty", 5, 0, 0666},
	}

	for _, d := range devices {
		dev := int(d.major*256 + d.minor)
		if err := syscall.Mknod(d.path, syscall.S_IFCHR|d.mode, dev); err != nil {
			return fmt.Errorf("mknod %s: %w", d.path, err)
		}
	}

	// /dev/ptmx symlink
	return os.Symlink("pts/ptmx", "/dev/ptmx")
}

func mountDevPts() error {
	os.MkdirAll("/dev/pts", 0755)
	if err := syscall.Mount("devpts", "/dev/pts", "devpts", 0, "newinstance,ptmxmode=0666"); err != nil {
		return fmt.Errorf("mount /dev/pts: %w", err)
	}

	os.MkdirAll("/dev/shm", 01777)
	if err := syscall.Mount("tmpfs", "/dev/shm", "tmpfs", syscall.MS_NOSUID|syscall.MS_NODEV, ""); err != nil {
		return fmt.Errorf("mount /dev/shm: %w", err)
	}

	return nil
}

func setupResolv() error {
	content := "nameserver 8.8.8.8\nnameserver 8.8.4.4\n"
	os.MkdirAll("/etc", 0755)
	if err := os.WriteFile("/etc/resolv.conf", []byte(content), 0644); err != nil {
		return fmt.Errorf("write /etc/resolv.conf: %w", err)
	}
	return nil
}
