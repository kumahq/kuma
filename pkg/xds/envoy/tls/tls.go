package tls

import (
	"fmt"
)

const (
	MeshCaResource       = "mesh_ca"
	IdentityCertResource = "identity_cert"
)

func MeshSpiffeIDPrefix(mesh string) string {
	return fmt.Sprintf("spiffe://%s/", mesh)
}

func ServiceSpiffeID(mesh string, service string) string {
	return fmt.Sprintf("spiffe://%s/%s", mesh, service)
}
