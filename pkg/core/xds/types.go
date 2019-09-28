package xds

import (
	"fmt"
	"github.com/Kong/kuma/pkg/core/logs"
	"github.com/Kong/kuma/pkg/core/permissions"
	"net"
	"strings"

	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/Kong/kuma/pkg/core/resources/model"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"

	"github.com/pkg/errors"
)

type ProxyId struct {
	Mesh      string
	Namespace string
	Name      string
}

func (id *ProxyId) String() string {
	return fmt.Sprintf("%s.%s.%s", id.Mesh, id.Name, id.Namespace)
}

type Proxy struct {
	Id                 ProxyId
	Dataplane          *mesh_core.DataplaneResource
	TrafficPermissions permissions.MatchedPermissions
	Logs               *logs.MatchedLogs
	OutboundTargets    map[string][]net.SRV
	Metadata           *DataplaneMetadata
}

func BuildProxyId(mesh, name string, more ...string) (*ProxyId, error) {
	id := strings.Join(append([]string{mesh, name}, more...), ".")
	return ParseProxyIdFromString(id)
}

func ParseProxyId(node *envoy_core.Node) (*ProxyId, error) {
	if node == nil {
		return nil, errors.Errorf("Envoy node must not be nil")
	}
	return ParseProxyIdFromString(node.Id)
}

func ParseProxyIdFromString(id string) (*ProxyId, error) {
	parts := strings.Split(id, ".")
	mesh := parts[0]
	if mesh == "" {
		return nil, errors.New("mesh must not be empty")
	}
	if len(parts) < 2 {
		return nil, errors.New("the name should be provided after the dot")
	}
	name := parts[1]
	if name == "" {
		return nil, errors.New("name must not be empty")
	}
	ns := core_model.DefaultNamespace
	if len(parts) == 3 {
		ns = parts[2]
	}
	if ns == "" {
		return nil, errors.New("namespace must not be empty")
	}
	return &ProxyId{
		Mesh:      mesh,
		Namespace: ns,
		Name:      name,
	}, nil
}

func (id *ProxyId) ToResourceKey() core_model.ResourceKey {
	return core_model.ResourceKey{
		Name:      id.Name,
		Namespace: id.Namespace,
		Mesh:      id.Mesh,
	}
}

func FromResourceKey(key core_model.ResourceKey) ProxyId {
	return ProxyId{
		Mesh:      key.Mesh,
		Namespace: key.Namespace,
		Name:      key.Name,
	}
}
