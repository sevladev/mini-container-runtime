package image

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const BaseDir = "/var/lib/minic/images"

type Manifest struct {
	Name      string    `json:"name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
}

func RootfsPath(name string) string {
	return filepath.Join(BaseDir, name, "rootfs")
}

func ManifestPath(name string) string {
	return filepath.Join(BaseDir, name, "manifest.json")
}

func Exists(name string) bool {
	_, err := os.Stat(RootfsPath(name))
	return err == nil
}

func HasManifest(name string) bool {
	_, err := os.Stat(ManifestPath(name))
	return err == nil
}

func SaveManifest(m Manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ManifestPath(m.Name), data, 0644)
}

func LoadManifest(name string) (Manifest, error) {
	data, err := os.ReadFile(ManifestPath(name))
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	return m, json.Unmarshal(data, &m)
}

func List() ([]Manifest, error) {
	entries, err := os.ReadDir(BaseDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var images []Manifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		m, err := LoadManifest(e.Name())
		if err != nil {
			continue
		}
		images = append(images, m)
	}
	return images, nil
}
