package issuer

import (
	"github.com/golang-jwt/jwt/v5"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/tokens"
)

type DataplaneIdentity struct {
	Name     string
	Mesh     string
	Tags     mesh_proto.MultiValueTagSet
	Type     mesh_proto.ProxyType
	Workload string
}

type DataplaneClaims struct {
	Name     string
	Mesh     string
	Tags     map[string][]string
	Type     string
	Workload string
	jwt.RegisteredClaims
}

func (d *DataplaneClaims) ID() string {
	return d.RegisteredClaims.ID
}

func (d *DataplaneClaims) SetRegisteredClaims(claims jwt.RegisteredClaims) {
	d.RegisteredClaims = claims
}

var _ tokens.Claims = &DataplaneClaims{}
