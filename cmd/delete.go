package cmd

import (
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/log"
	"github.com/spf13/cobra"
)

func NewCmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete all service blocks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cloudrun.NewClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			err = client.DeleteAll(ctx)
			if err != nil {
				return err
			}

			log.Info(ctx, "deleted all service blocks")
			return nil
		},
	}

	return cmd
}
