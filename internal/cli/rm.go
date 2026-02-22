package cli

import (
	"github.com/sevladev/minic/internal/container"
	"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm CONTAINER",
		Short: "Remove a stopped container",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return container.Remove(args[0])
		},
	}
}
