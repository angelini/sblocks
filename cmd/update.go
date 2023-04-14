package cmd

import (
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/spf13/cobra"
)

func NewCmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service blocks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cloudrun.NewClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			// TODO

			return nil
		},
	}

	return cmd
}
