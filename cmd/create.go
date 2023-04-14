package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/angelini/sblocks/internal/log"
	"github.com/angelini/sblocks/pkg/cloudrun"
	cr "github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/core"
	rt "github.com/angelini/sblocks/pkg/runtime"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdCreate() *cobra.Command {
	var (
		runtime string
		size    int
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create service block",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			client, err := cr.NewClient(ctx, os.Getenv("GCP_PROJECT"), os.Getenv("GCP_REGION"))
			if err != nil {
				return err
			}
			defer client.Close()

			labels := map[string]string{
				"sb_runtime": runtime,
			}

			runtime := rt.NewRuntime(runtime, true, size, labels, core.Revision{
				Name:           "1",
				MinScale:       1,
				MaxScale:       2,
				MaxConcurrency: 50,
				Timeout:        time.Minute,
				Containers: map[string]core.Container{
					"deno": {Name: "deno", Image: os.Getenv("DENO_IMAGE")},
				},
			})

			log.Info(ctx, "created service block", zap.Int("size", size))

			block.CreateRevision(ctx, client, &cloudrun.Revision{
				Name:           "2",
				MinScale:       1,
				MaxScale:       2,
				MaxConcurrency: 50,
				Timeout:        time.Minute,
				Containers: map[string]cloudrun.Container{
					"deno": {Name: "deno", Image: os.Getenv("DENO_IMAGE")},
				},
			})

			fmt.Println()
			for _, line := range runtime.Display() {
				fmt.Println(line)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&runtime, "runtime", "r", "", "Name of the runtime that will be created")
	cmd.PersistentFlags().IntVarP(&size, "size", "s", 10, "Free size")

	cmd.MarkPersistentFlagRequired("runtime")

	return cmd
}
