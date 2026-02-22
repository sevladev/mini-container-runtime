package cli

import (
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/sevladev/minic/internal/container"
	"github.com/spf13/cobra"
)

func newPsCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List containers",
		RunE: func(cmd *cobra.Command, args []string) error {
			containers, err := container.ListStates()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "CONTAINER ID\tIMAGE\tCOMMAND\tSTATUS\tIP")

			for _, c := range containers {
				if !all && c.Status != container.StateRunning {
					continue
				}
				command := strings.Join(c.Command, " ")
				if len(command) > 30 {
					command = command[:27] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					c.ID, c.Image, command, c.Status, c.IP)
			}

			return w.Flush()
		},
	}

	cmd.Flags().BoolVarP(&all, "all", "a", false, "Show all containers (default shows just running)")

	return cmd
}
