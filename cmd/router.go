package cmd

import (
	"github.com/angelini/sblocks/internal/log"
	"github.com/angelini/sblocks/pkg/router"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCmdRouter() *cobra.Command {
	var (
		port int
	)

	cmd := &cobra.Command{
		Use:   "router",
		Short: "Router HTTP service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()

			proxy, err := router.NewProxy(port)
			if err != nil {
				return err
			}

			log.Info(ctx, "start router", zap.Int("port", port))
			return proxy.Start(ctx)
		},
	}

	cmd.PersistentFlags().IntVarP(&port, "port", "p", 5021, "Listen port")

	return cmd
}
