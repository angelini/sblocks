package cmd

import (
	"os"
	"time"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/log"
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

			client, err := cloudrun.NewCloudRunClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			block := cloudrun.NewServiceBlock("example", size, map[string]string{})

			err = block.Create(ctx, client, &cloudrun.Revision{
				Name:           "1",
				MinScale:       1,
				MaxScale:       2,
				MaxConcurrency: 50,
				Timeout:        time.Minute,
				Containers: map[string]cloudrun.Container{
					"deno": {Name: "deno", Image: os.Getenv("DENO_IMAGE")},
				},
			})
			if err != nil {
				return err
			}

			log.Info(ctx, "created service block", zap.Int("size", size))
			return nil
		},
	}

	cmd.PersistentFlags().IntVarP(&size, "size", "s", 10, "Size of the service block")

	return cmd
}
