package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/maps"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func NewCmdRouter() *cobra.Command {
	var (
		environment string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List service blocks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			etcd, err := clientv3.New(clientv3.Config{
				Endpoints:   []string{"localhost:2379", "localhost:22379", "localhost:32379"},
				DialTimeout: 5 * time.Second,
			})
			if err != nil {
				return err
			}
			defer etcd.Close()

			client, err := cloudrun.NewClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			blocks, err := cloudrun.LoadServiceBlocks(ctx, client, environment)
			if err != nil {
				return err
			}

			for idx, block := range maps.SortedValues(blocks) {
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
