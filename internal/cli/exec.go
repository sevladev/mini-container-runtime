package cli

import (
	"fmt"

	"github.com/sevladev/minic/internal/container"
	"github.com/sevladev/minic/internal/namespace"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "exec CONTAINER COMMAND [ARG...]",
		Short: "Execute a command in a running container",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			meta, err := container.FindByPrefix(args[0])
			if err != nil {
				return err
			}

			if meta.Status != container.StateRunning {
				return fmt.Errorf("container %s is not running", meta.ID)
			}

			return namespace.EnterAndExec(meta.PID, args[1:])
		},
	}
}
