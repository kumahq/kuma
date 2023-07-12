package etcd

import (
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/pkg/errors"
	"strings"
)

type etcdKey struct {
	prefix string
	typ    core_model.ResourceType
	mesh   string
	name   string
}

func (e *etcdKey) GetPrefix() string {
	return e.prefix
}
func (e etcdKey) GetResourceType() core_model.ResourceType {
	return e.typ
}
func (e etcdKey) GetMesh() string {
	return e.mesh
}

func (e etcdKey) GetName() string {
	return e.name
}

func NewEtcdKey(prefix string, typ core_model.ResourceType, mesh, name string) etcdKey {
	return etcdKey{
		prefix: prefix,
		typ:    typ,
		mesh:   mesh,
		name:   name,
	}
}

func WithEtcdKey(key string) (etcdKey, error) {
	split := strings.Split(key, "/")
	if len(split) != 5 {
		return etcdKey{}, errors.New("key format error ")
	}
	return etcdKey{
		prefix: split[1],
		typ:    core_model.ResourceType(split[2]),
		mesh:   split[3],
		name:   split[4],
	}, nil
}

func (k etcdKey) String() string {
	builder := strings.Builder{}

	builder.WriteString("/")
	if len(k.prefix) != 0 {
		builder.WriteString(k.prefix)
	}

	builder.WriteString("/")
	if len(k.typ) != 0 {
		builder.WriteString(string(k.typ))
	}

	builder.WriteString("/")
	if len(k.mesh) != 0 {
		builder.WriteString(k.mesh)
	}

	builder.WriteString("/")
	if len(k.name) != 0 {
		builder.WriteString(k.name)
	}
	return builder.String()
}
