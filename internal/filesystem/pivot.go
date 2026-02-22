//go:build linux

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

func PivotRoot(rootfs string) error {
	// TODO: pivot_root requires the new root to be a mount point
	if err := syscall.Mount(rootfs, rootfs, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("bind mount rootfs: %w", err)
	}

	putOld := filepath.Join(rootfs, ".pivot_root")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("mkdir put_old: %w", err)
	}

	if err := syscall.PivotRoot(rootfs, putOld); err != nil {
		return fmt.Errorf("pivot_root: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir /: %w", err)
	}

	if err := syscall.Unmount("/.pivot_root", syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount put_old: %w", err)
	}

	return os.RemoveAll("/.pivot_root")
}
