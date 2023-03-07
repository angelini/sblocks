package cmd

import (
	"context"

	"github.com/angelini/sblocks/pkg/log"
	"github.com/spf13/cobra"
)

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "sblocks",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			ctx, err := log.Init(cmd.Context())
			if err != nil {
				return err
			}

			cmd.SetContext(ctx)

			return nil
		},
	}

	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdDelete())
	cmd.AddCommand(NewCmdUpdate())

	return cmd
}

func Execute() error {
	ctx := context.Background()
	return NewCmdRoot().ExecuteContext(ctx)
}
