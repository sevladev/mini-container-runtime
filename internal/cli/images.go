package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/sevladev/minic/internal/image"
	"github.com/spf13/cobra"
)

func newImagesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "images",
		Short: "List local images",
		RunE: func(cmd *cobra.Command, args []string) error {
			images, err := image.List()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "IMAGE\tSIZE\tCREATED")

			for _, img := range images {
				size := fmt.Sprintf("%.1f MB", float64(img.Size)/1024/1024)
				created := img.CreatedAt.Format("2006-01-02 15:04")
				fmt.Fprintf(w, "%s\t%s\t%s\n", img.Name, size, created)
			}

			return w.Flush()
		},
	}
}
