//go:build !linux

package container

import "fmt"

var errLinux = fmt.Errorf("minic requires Linux (namespaces, cgroups, pivot_root are Linux-only)")

func Run(cfg Config) error  { return errLinux }
func Stop(id string) error  { return errLinux }
func Remove(id string) error { return errLinux }
