package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ps",
		Short: "List running containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("CONTAINER ID\tIMAGE\tCOMMAND\tSTATUS\tNAME")
			return nil
		},
	}
}
