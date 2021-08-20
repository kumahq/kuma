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

var defaultRetry = &mesh_proto.Retry{
	Sources: []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		},
	},
	Destinations: []*mesh_proto.Selector{
		{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		},
	},
	Conf: &mesh_proto.Retry_Conf{
		Http: &mesh_proto.Retry_Conf_Http{
			NumRetries:    util_proto.UInt32(5),
			PerTryTimeout: util_proto.Duration(16 * time.Second),
			BackOff: &mesh_proto.Retry_Conf_BackOff{
				BaseInterval: util_proto.Duration(25 * time.Millisecond),
				MaxInterval:  util_proto.Duration(250 * time.Millisecond),
			},
		},
		Tcp: &mesh_proto.Retry_Conf_Tcp{
			MaxConnectAttempts: 5,
		},
		Grpc: &mesh_proto.Retry_Conf_Grpc{
			NumRetries:    util_proto.UInt32(5),
			PerTryTimeout: util_proto.Duration(16 * time.Second),
			BackOff: &mesh_proto.Retry_Conf_BackOff{
				BaseInterval: util_proto.Duration(25 * time.Millisecond),
				MaxInterval:  util_proto.Duration(250 * time.Millisecond),
			},
		},
	},
}

// Retry needs to contain mesh name inside it. Otherwise if the name is the
//  same (ex. "retry-all") creating new mesh would fail because there is already
//  resource of name "retry-all" which is unique key on K8S
func defaultRetryKey(meshName string) model.ResourceKey {
	return model.ResourceKey{
		Mesh: meshName,
		Name: fmt.Sprintf("retry-all-%s", meshName),
	}
}

func ensureDefaultRetry(resManager manager.ResourceManager, meshName string) (error, bool) {
	retry := &core_mesh.RetryResource{
		Spec: defaultRetry,
	}

	key := defaultRetryKey(meshName)

	err := resManager.Get(context.Background(), retry, store.GetBy(key))
	if err == nil {
		return err, false
	}

	if !store.IsResourceNotFound(err) {
		return errors.Wrap(err, "could not retrieve a resource"), false
	}

	if err := resManager.Create(context.Background(), retry, store.CreateBy(key)); err != nil {
		return errors.Wrap(err, "could not create a resource"), false
	}

	return nil, true
}
