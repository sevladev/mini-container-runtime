package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "images",
		Short: "List local images",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("IMAGE\tSIZE\tCREATED")
			return nil
		},
	}
}
