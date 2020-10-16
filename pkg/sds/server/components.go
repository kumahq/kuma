package server

import (
	core_runtime "github.com/kumahq/kuma/pkg/core/runtime"
	sds_provider "github.com/kumahq/kuma/pkg/sds/provider"
	ca_sds_provider "github.com/kumahq/kuma/pkg/sds/provider/ca"
	identity_sds_provider "github.com/kumahq/kuma/pkg/sds/provider/identity"
)

func DefaultMeshCaProvider(rt core_runtime.Runtime) sds_provider.SecretProvider {
	return ca_sds_provider.New(rt.ResourceManager(), rt.CaManagers())
}

func DefaultIdentityCertProvider(rt core_runtime.Runtime) sds_provider.SecretProvider {
	return identity_sds_provider.New(rt.ResourceManager(), rt.CaManagers())
}
