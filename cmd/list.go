package cmd

import (
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/log"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service blocks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cloudrun.NewCloudRunClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			results, err := client.List(ctx)
			if err != nil {
				return err
			}

			for _, result := range results {
				log.Info(ctx, "service", zap.String("name", result))
			}

			return nil
		},
	}

	return cmd
}
