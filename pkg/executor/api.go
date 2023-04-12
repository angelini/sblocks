package executor

import (
	"github.com/angelini/sblocks/pkg/cloudrun"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type ExecutorApi struct {
	etcd     *clientv3.Client
	cloudrun *cloudrun.Client
}

func NewExecutorApi(etcd *clientv3.Client, cr *cloudrun.Client) *ExecutorApi {
	return &ExecutorApi{
		etcd:     etcd,
		cloudrun: cr,
	}
}
