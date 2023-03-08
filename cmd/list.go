package cmd

import (
	"fmt"
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/spf13/cobra"
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

			block, err := cloudrun.LoadServiceBlock(ctx, client, "example")
			if err != nil {
				return err
			}

			for _, line := range block.Display() {
				fmt.Println(line)
			}

			return nil
		},
	}

	return cmd
}
