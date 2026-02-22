//go:build !linux

package container

import "fmt"

func Init(args []string) error {
	return fmt.Errorf("minic requires Linux")
}
