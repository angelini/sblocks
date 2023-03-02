package cmd

import (
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdCreate() *cobra.Command {
	var (
		size int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create service block",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			log := ctx.Value(logKey).(*zap.Logger)

			client, err := cloudrun.NewCloudRunClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			block := cloudrun.NewServiceBlock(cloudrun.Service{
				RootName: "example",
				Labels:   map[string]string{},
				Containers: []cloudrun.Container{
					{Name: "deno", Image: os.Getenv("DENO_IMAGE")},
				},
			}, size)

			err = block.Create(ctx, client, "1")
			if err != nil {
				return err
			}
			log.Info("created service block", zap.Int("size", size))

			return nil
		},
	}

	cmd.PersistentFlags().IntVarP(&size, "size", "s", 10, "Size of the service block")

	return cmd
}
