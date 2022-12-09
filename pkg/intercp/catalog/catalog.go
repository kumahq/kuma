package catalog

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type Instance struct {
	Id          string `json:"id"`
	Address     string `json:"address"`
	InterCpPort uint16 `json:"interCpPort"`
	Leader      bool   `json:"leader"`
}

func (i Instance) InterCpURL() string {
	return fmt.Sprintf("grpcs://%s:%d", i.Address, i.InterCpPort)
}

type Catalog interface {
	Instances(context.Context) ([]Instance, error)
	Replace(context.Context, []Instance) (bool, error)
}

var ErrNoLeader = errors.New("leader not found")

func Leader(ctx context.Context, catalog Catalog) (Instance, error) {
	instances, err := catalog.Instances(ctx)
	if err != nil {
		return Instance{}, err
	}
	for _, instance := range instances {
		if instance.Leader {
			return instance, nil
		}
	}
	return Instance{}, ErrNoLeader
}
