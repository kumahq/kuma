package envoy

import (
	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/xds/envoy/names"
	xds_tls "github.com/kumahq/kuma/pkg/xds/envoy/tls"
)

type identityCertRequest struct {
	meshName string
}

func (r identityCertRequest) Name() string {
	return names.GetSecretName(xds_tls.IdentityCertResource, "secret", r.meshName)
}

func (r identityCertRequest) MeshName() string {
	return r.meshName
}

type caRequest struct {
	meshName string
}

type allInOneCaRequest struct {
	meshNames []string
}

func (r caRequest) Name() string {
	return names.GetSecretName(xds_tls.MeshCaResource, "secret", r.meshName)
}

func (r caRequest) MeshName() []string {
	return []string{r.meshName}
}

func (r allInOneCaRequest) Name() string {
	return names.GetSecretName(xds_tls.MeshCaResource, "secret", "all")
}

func (r allInOneCaRequest) MeshName() []string {
	return r.meshNames
}

type secretsTracker struct {
	ownMesh   string
	allMeshes []string

	identity bool
	meshes   map[string]struct{}
	allInOne bool
}

func NewSecretsTracker(ownMesh string, allMeshes []string) core_xds.SecretsTracker {
	return &secretsTracker{
		ownMesh:   ownMesh,
		allMeshes: allMeshes,

		meshes: map[string]struct{}{},
	}
}

func (st *secretsTracker) RequestIdentityCert() core_xds.IdentityCertRequest {
	st.identity = true
	return &identityCertRequest{
		meshName: st.ownMesh,
	}
}

func (st *secretsTracker) RequestCa(mesh string) core_xds.CaRequest {
	st.meshes[mesh] = struct{}{}
	return &caRequest{
		meshName: mesh,
	}
}

func (st *secretsTracker) RequestAllInOneCa() core_xds.CaRequest {
	st.allInOne = true
	return &allInOneCaRequest{
		meshNames: st.allMeshes,
	}
}

func (st *secretsTracker) UsedIdentity() bool {
	return st.identity
}

func (st *secretsTracker) UsedCas() map[string]struct{} {
	return st.meshes
}

func (st *secretsTracker) UsedAllInOne() bool {
	return st.allInOne
}
