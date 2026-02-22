//go:build linux

package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

const containerBaseDir = "/var/lib/minic/containers"

type OverlayDirs struct {
	ContainerDir string
	Lower        string
	Upper        string
	Work         string
	Merged       string
}

func SetupOverlay(containerID string, imageRootfs string) (*OverlayDirs, error) {
	dir := filepath.Join(containerBaseDir, containerID)

	dirs := &OverlayDirs{
		ContainerDir: dir,
		Lower:        imageRootfs,
		Upper:        filepath.Join(dir, "upper"),
		Work:         filepath.Join(dir, "work"),
		Merged:       filepath.Join(dir, "merged"),
	}

	for _, d := range []string{dirs.Upper, dirs.Work, dirs.Merged} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("create dir %s: %w", d, err)
		}
	}

	opts := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", dirs.Lower, dirs.Upper, dirs.Work)
	if err := syscall.Mount("overlay", dirs.Merged, "overlay", 0, opts); err != nil {
		return nil, fmt.Errorf("mount overlay: %w", err)
	}

	return dirs, nil
}

func RemoveOverlay(dirs *OverlayDirs) error {
	if err := syscall.Unmount(dirs.Merged, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount overlay: %w", err)
	}
	return os.RemoveAll(dirs.ContainerDir)
}
