package etcd

import "github.com/pkg/errors"

// Etcd store configuration
type EtcdConfig struct {
	Endpoints []string `json:"endpoints"`
}

func (e *EtcdConfig) Validate() error {
	if len(e.Endpoints) < 1 {
		return errors.New("Endpoints should not be empty")
	}
	return nil
}
