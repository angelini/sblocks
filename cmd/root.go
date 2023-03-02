package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type configKey string

var logKey = configKey("log")

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "sblocks",
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			config := zap.NewDevelopmentConfig()
			log, err := config.Build()
			if err != nil {
				return fmt.Errorf("cannot build zap logger: %w", err)
			}

			ctx := cmd.Context()
			cmd.SetContext(context.WithValue(ctx, logKey, log))

			return nil
		},
	}

	cmd.AddCommand(NewCmdCreate())
	cmd.AddCommand(NewCmdList())
	cmd.AddCommand(NewCmdDelete())

	return cmd
}

func Execute() error {
	ctx := context.Background()
	return NewCmdRoot().ExecuteContext(ctx)
}
