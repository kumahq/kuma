package mesh

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultTimeout = mesh_proto.Timeout{
	Sources: []*mesh_proto.Selector{{
		Match: mesh_proto.MatchAnyService(),
	}},
	Destinations: []*mesh_proto.Selector{{
		Match: mesh_proto.MatchAnyService(),
	}},
	Conf: &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(5 * time.Second),
		Tcp: &mesh_proto.Timeout_Conf_Tcp{
			IdleTimeout: util_proto.Duration(1 * time.Hour),
		},
		Http: &mesh_proto.Timeout_Conf_Http{
			IdleTimeout:    util_proto.Duration(1 * time.Hour),
			RequestTimeout: util_proto.Duration(15 * time.Second),
		},
		Grpc: &mesh_proto.Timeout_Conf_Grpc{
			StreamIdleTimeout: util_proto.Duration(5 * time.Minute),
		},
	},
}

// Timeout needs to contain mesh name inside it. Otherwise if the name is the same (ex. "allow-all") creating new mesh would fail because there is already resource of name "allow-all" which is unique key on K8S
func defaultTimeoutKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("timeout-all-%s", meshName),
	}
}

func ensureDefaultTimeout(resManager manager.ResourceManager, meshName string) (err error, created bool) {
	timeout := &core_mesh.TimeoutResource{
		Spec: &defaultTimeout,
	}
	key := defaultTimeoutKey(meshName)
	err = resManager.Get(context.Background(), timeout, store.GetBy(key))
	if err == nil {
		return nil, false
	}
	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}
	if err := resManager.Create(context.Background(), timeout, store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}
	return nil, true
}
