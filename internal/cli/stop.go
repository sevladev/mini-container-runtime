package cli

import (
	"github.com/sevladev/minic/internal/container"
	"github.com/spf13/cobra"
)

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop CONTAINER",
		Short: "Stop a running container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.Stop(args[0])
		},
	}
}
