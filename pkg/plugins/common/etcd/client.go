package etcd

import (
	clientv3 "go.etcd.io/etcd/client/v3"

	config "github.com/kumahq/kuma/pkg/config/plugins/resources/etcd"
)

func NewClient(config *config.EtcdConfig) (*clientv3.Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:            config.Endpoints,
		AutoSyncInterval:     config.AutoSyncInterval,
		DialTimeout:          config.DialTimeout,
		DialKeepAliveTime:    config.DialKeepAliveTime,
		DialKeepAliveTimeout: config.DialKeepAliveTimeout,
		MaxCallSendMsgSize:   config.MaxCallSendMsgSize,
		MaxCallRecvMsgSize:   config.MaxCallRecvMsgSize,
		Username:             config.Username,
		Password:             config.Password,
	})
	return client, err
}
