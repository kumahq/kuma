package issuer

import (
	"github.com/golang-jwt/jwt/v4"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/tokens"
)

func DataplaneTokenSigningKeyPrefix(mesh string) string {
	return "dataplane-token-signing-key-" + mesh
}

func DataplaneTokenRevocationsSecretKey(mesh string) core_model.ResourceKey {
	return core_model.ResourceKey{
		Name: "dataplane-token-revocations-" + mesh,
		Mesh: mesh,
	}
}

type DataplaneIdentity struct {
	Name string
	Mesh string
	Tags mesh_proto.MultiValueTagSet
	Type mesh_proto.ProxyType
}

type DataplaneClaims struct {
	Name string
	Mesh string
	Tags map[string][]string
	Type string
	jwt.RegisteredClaims
}

func (d *DataplaneClaims) ID() string {
	return d.RegisteredClaims.ID
}

func (d *DataplaneClaims) KeyIDFallback() (int, error) {
	return 0, nil
}

func (d *DataplaneClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	d.RegisteredClaims = claims
}

var _ tokens.Claims = &DataplaneClaims{}
