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

func KumaID(tagName, tagValue string) string {
	return fmt.Sprintf("kuma://%s/%s", tagName, tagValue)
}
