package catalog

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

type Instance struct {
	Id          string `json:"id"`
	Address     string `json:"address"`
	InterCpPort uint16 `json:"interCpPort"`
	Leader      bool   `json:"leader"`
}

func (i Instance) InterCpURL() string {
	return fmt.Sprintf("grpcs://%s", net.JoinHostPort(i.Address, strconv.Itoa(int(i.InterCpPort))))
}

type Reader interface {
	Instances(context.Context) ([]Instance, error)
}

type Catalog interface {
	Reader
	Replace(context.Context, []Instance) (bool, error)
	ReplaceLeader(context.Context, Instance) error
	DropLeader(context.Context, Instance) error
}

var (
	ErrNoLeader         = errors.New("leader not found")
	ErrInstanceNotFound = errors.New("instance not found")
)

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

func InstanceOfID(ctx context.Context, catalog Catalog, id string) (Instance, error) {
	instances, err := catalog.Instances(ctx)
	if err != nil {
		return Instance{}, err
	}
	for _, instance := range instances {
		if instance.Id == id {
			return instance, nil
		}
	}
	return Instance{}, ErrInstanceNotFound
}

type InstancesByID []Instance

func (a InstancesByID) Len() int      { return len(a) }
func (a InstancesByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a InstancesByID) Less(i, j int) bool {
	return a[i].Id < a[j].Id
}
