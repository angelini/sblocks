package cmd

import (
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete all service blocks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			log := ctx.Value(logKey).(*zap.Logger)

			client, err := cloudrun.NewCloudRunClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			err = client.DeleteAll(ctx)
			if err != nil {
				return err
			}

			log.Info("deleted all service blocks")
			return nil
		},
	}

	return cmd
}
