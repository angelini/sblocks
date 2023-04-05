package cmd

import (
	"fmt"
	"os"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/maps"
	"github.com/spf13/cobra"
)

func NewCmdList() *cobra.Command {
	var (
		environment string
	)

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

			blocks, err := cloudrun.LoadServiceBlocks(ctx, client, environment)
			if err != nil {
				return err
			}

			for idx, block := range maps.SortedRange(blocks) {
				if idx != 0 {
					fmt.Println("---------------")
				}
				for _, line := range block.Display() {
					fmt.Println(line)
				}
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&environment, "environment", "e", "", "Name of the environment that the block will be added to")

	cmd.MarkPersistentFlagRequired("environment")

	return cmd
}
