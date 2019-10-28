package ca

import (
	"context"

	"github.com/pkg/errors"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	builtin_ca "github.com/Kong/kuma/pkg/core/ca/builtin"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	core_store "github.com/Kong/kuma/pkg/core/resources/store"
	sds_auth "github.com/Kong/kuma/pkg/sds/auth"
	sds_provider "github.com/Kong/kuma/pkg/sds/provider"
)

func New(resourceManager core_manager.ResourceManager, builtinCaManager builtin_ca.BuiltinCaManager) sds_provider.SecretProvider {
	return &meshCaProvider{
		resourceManager:  resourceManager,
		builtinCaManager: builtinCaManager,
	}
}

type meshCaProvider struct {
	resourceManager  core_manager.ResourceManager
	builtinCaManager builtin_ca.BuiltinCaManager
}

func (s *meshCaProvider) RequiresIdentity() bool {
	return false
}

func (s *meshCaProvider) Get(ctx context.Context, resource string, requestor sds_auth.Identity) (sds_provider.Secret, error) {
	meshName := requestor.Mesh
	list := &core_mesh.MeshResourceList{}
	if err := s.resourceManager.List(ctx, list, core_store.ListByMesh(meshName)); err != nil {
		return nil, errors.Wrapf(err, "failed to find a Mesh %q", meshName)
	}
	if len(list.Items) == 0 {
		return nil, errors.Errorf("there is no Mesh %q", meshName)
	}
	if len(list.Items) != 1 {
		return nil, errors.Errorf("there are multiple Meshes named %q", meshName)
	}
	mesh := list.Items[0]
	switch mesh.Spec.GetMtls().GetCa().GetType().(type) {
	case *mesh_proto.CertificateAuthority_Builtin_:
		rootCerts, err := s.builtinCaManager.GetRootCerts(ctx, mesh.Meta.GetName())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to retrieve Root Certificates of a given Builtin CA")
		}
		return &MeshCaSecret{
			PemCerts: rootCerts,
		}, nil
	default:
		return nil, errors.Errorf("Mesh %q has unsupported CA type", meshName)
	}
}
