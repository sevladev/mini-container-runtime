package cli

import (
	"github.com/sevladev/minic/internal/image"
	"github.com/spf13/cobra"
)

func newPullCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pull IMAGE",
		Short: "Download a rootfs image",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return image.Pull(args[0])
		},
	}
}
