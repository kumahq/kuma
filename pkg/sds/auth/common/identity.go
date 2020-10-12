package common

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	sds_auth "github.com/kumahq/kuma/pkg/sds/auth"
)

type DataplaneResolver func(context.Context, core_xds.ProxyId) (*core_mesh.DataplaneResource, error)

func GetDataplaneIdentity(dataplane *core_mesh.DataplaneResource) (sds_auth.Identity, error) {
	services := dataplane.Spec.TagSet().Values(mesh_proto.ServiceTag)
	if len(services) == 0 {
		return sds_auth.Identity{}, errors.Errorf("Dataplane has no services associated with it")
	}
	return sds_auth.Identity{
		Mesh:     dataplane.Meta.GetMesh(),
		Services: services,
	}, nil
}
