package executor

import (
	"context"

	pb "github.com/angelini/sblocks/internal/executorpb"
	"github.com/angelini/sblocks/internal/log"
	"github.com/angelini/sblocks/pkg/cloudrun"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

func NewServer(ctx context.Context, etcd *clientv3.Client, cr *cloudrun.Client) *grpc.Server {
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(log.GetLogger(ctx)),
			),
		),
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(log.GetLogger(ctx)),
			),
		),
	)

	api := NewExecutorApi(etcd, cr)
	pb.RegisterExecutorServer(grpcServer, api)

	return grpcServer
}
