package xds

import (
	"strings"

	core_model "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/pkg/errors"
)

func ParseDataplaneId(node *envoy_core.Node) (*core_model.ResourceKey, error) {
	if node == nil {
		return nil, errors.Errorf("Envoy node must not be nil")
	}
	parts := strings.Split(node.Id, ".")
	name := parts[0]
	ns := core_model.DefaultNamespace
	if 1 < len(parts) {
		ns = parts[1]
	}
	mesh := core_model.DefaultMesh
	return &core_model.ResourceKey{
		Mesh:      mesh,
		Namespace: ns,
		Name:      name,
	}, nil
}
