package cli

import (
	"fmt"

	"github.com/sevladev/minic/internal/container"
	"github.com/sevladev/minic/pkg/units"
	"github.com/spf13/cobra"
)

type runOptions struct {
	memory   string
	cpus     float64
	pids     int
	name     string
	hostname string
	net      string
	detach   bool
	volumes  []string
}

func newRunCmd() *cobra.Command {
	opts := &runOptions{}

	cmd := &cobra.Command{
		Use:   "run [flags] IMAGE COMMAND [ARG...]",
		Short: "Create and run a new container",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var memBytes int64
			if opts.memory != "" {
				var err error
				memBytes, err = units.ParseMemory(opts.memory)
				if err != nil {
					return fmt.Errorf("invalid --memory: %w", err)
				}
			}

			cfg := container.Config{
				Image:    args[0],
				Command:  args[1:],
				Hostname: opts.hostname,
				NetMode:  opts.net,
				Volumes:  opts.volumes,
				Resources: container.ResourceLimits{
					MemoryBytes: memBytes,
					CPUQuota:    opts.cpus,
					PidsMax:     opts.pids,
				},
			}

			return container.Run(cfg)
		},
	}

	cmd.Flags().StringVarP(&opts.memory, "memory", "m", "", "Memory limit (e.g. 100m, 1g)")
	cmd.Flags().Float64Var(&opts.cpus, "cpus", 0, "CPU limit (e.g. 0.5, 2.0)")
	cmd.Flags().IntVar(&opts.pids, "pids", 0, "Max number of PIDs")
	cmd.Flags().StringVar(&opts.name, "name", "", "Container name")
	cmd.Flags().StringVar(&opts.hostname, "hostname", "", "Container hostname")
	cmd.Flags().StringVar(&opts.net, "net", "bridge", "Network mode: bridge or none")
	cmd.Flags().BoolVarP(&opts.detach, "detach", "d", false, "Run in background")
	cmd.Flags().StringSliceVarP(&opts.volumes, "volume", "v", nil, "Bind mount (host:container)")

	return cmd
}
