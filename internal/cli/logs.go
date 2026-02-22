package cli

import (
	"fmt"
	"os"

	"github.com/sevladev/minic/internal/container"
	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs CONTAINER",
		Short: "Show container logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			meta, err := container.FindByPrefix(args[0])
			if err != nil {
				return err
			}

			stdout, err := os.ReadFile(container.StdoutLogPath(meta.ID))
			if err == nil && len(stdout) > 0 {
				fmt.Print(string(stdout))
			}

			stderr, err := os.ReadFile(container.StderrLogPath(meta.ID))
			if err == nil && len(stderr) > 0 {
				fmt.Fprint(os.Stderr, string(stderr))
			}

			return nil
		},
	}
}
