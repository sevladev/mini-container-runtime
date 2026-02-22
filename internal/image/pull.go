package image

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

var registry = map[string]string{
	"alpine": "https://dl-cdn.alpinelinux.org/alpine/v3.21/releases/x86_64/alpine-minirootfs-3.21.3-x86_64.tar.gz",
}

func Pull(name string) error {
	url, ok := registry[name]
	if !ok {
		return fmt.Errorf("unknown image %q (available: alpine)", name)
	}

	if Exists(name) && HasManifest(name) {
		fmt.Printf("Image %q already exists\n", name)
		return nil
	}

	rootfs := RootfsPath(name)
	if err := os.MkdirAll(rootfs, 0755); err != nil {
		return fmt.Errorf("create rootfs dir: %w", err)
	}

	fmt.Printf("Pulling %s...\n", name)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "minic-pull-*.tar.gz")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	size, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}

	cmd := exec.Command("tar", "-xzf", tmpFile.Name(), "-C", rootfs)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("extract: %s: %w", string(out), err)
	}

	manifest := Manifest{
		Name:      name,
		Size:      size,
		CreatedAt: time.Now(),
	}
	if err := SaveManifest(manifest); err != nil {
		return fmt.Errorf("save manifest: %w", err)
	}

	fmt.Printf("Pulled %s (%.1f MB)\n", name, float64(size)/1024/1024)
	return nil
}
