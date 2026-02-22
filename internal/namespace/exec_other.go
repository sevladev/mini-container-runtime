//go:build !linux

package namespace

import "fmt"

func EnterAndExec(pid int, command []string) error {
	return fmt.Errorf("minic requires Linux")
}
