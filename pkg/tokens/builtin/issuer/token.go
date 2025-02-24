package issuer

import (
	"github.com/golang-jwt/jwt/v4"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/tokens"
)

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

func (d *DataplaneClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	d.RegisteredClaims = claims
}

var _ tokens.Claims = &DataplaneClaims{}
