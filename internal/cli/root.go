package cli

import (
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minic",
		Short: "A minimal container runtime",
		Long:  "minic is a minimal container runtime that isolates processes using Linux namespaces, cgroups v2, and pivot_root.",
	}

	cmd.AddCommand(
		newRunCmd(),
		newExecCmd(),
		newPsCmd(),
		newStopCmd(),
		newRmCmd(),
		newImagesCmd(),
		newPullCmd(),
		newLogsCmd(),
	)

	return cmd
}
