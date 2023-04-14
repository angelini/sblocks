package executor

import (
	"context"

	pb "github.com/angelini/sblocks/internal/executorpb"
	"github.com/angelini/sblocks/pkg/cloudrun"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ExecutorApi struct {
	pb.UnimplementedExecutorServer

	etcd     *clientv3.Client
	cloudrun *cloudrun.Client
}

func NewExecutorApi(etcd *clientv3.Client, cr *cloudrun.Client) *ExecutorApi {
	return &ExecutorApi{
		etcd:     etcd,
		cloudrun: cr,
	}
}

func (a *ExecutorApi) GetService(ctx context.Context, req *pb.GetServiceRequest) (*pb.GetServiceResponse, error) {
	return &pb.GetServiceResponse{
		Runtime: req.Runtime,
		Id:      req.Id,
		Uri:     "",
		State:   pb.State_READY,
	}, nil
}

func (a *ExecutorApi) UpdateRuntime(ctx context.Context, req *pb.UpdateRuntimeRequest) (*pb.UpdateRuntimeResponse, error) {
	return &pb.UpdateRuntimeResponse{
		Output: "",
	}, nil
}
