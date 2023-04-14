package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/angelini/sblocks/internal/log"
	"github.com/angelini/sblocks/pkg/cloudrun"
	"github.com/angelini/sblocks/pkg/executor"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
)

func NewCmdExecutor() *cobra.Command {
	var (
		port int
	)

	cmd := &cobra.Command{
		Use:   "executor",
		Short: "Executor GRPC service",
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

			socket, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
			if err != nil {
				return fmt.Errorf("failed to listen on TCP port %d: %w", port, err)
			}

			server := executor.NewServer(ctx, etcd, client)

			osSignals := make(chan os.Signal, 1)
			signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
			go func() {
				<-osSignals
				server.GracefulStop()
			}()

			log.Info(ctx, "start executor", zap.Int("port", port))
			return server.Serve(socket)
		},
	}

	cmd.PersistentFlags().IntVarP(&port, "port", "p", 5020, "Listen port")

	cmd.MarkPersistentFlagRequired("environment")

	return cmd
}
