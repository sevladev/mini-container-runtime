//go:build !linux

package container

import "fmt"

func Run(cfg Config) error {
	return fmt.Errorf("minic requires Linux (namespaces, cgroups, pivot_root are Linux-only)")
}
