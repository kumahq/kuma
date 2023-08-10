package etcd

import (
	"strings"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type storeType string

func (s storeType) String() string {
	return string(s)
}

type EtcdKey interface {
	String() string
	Prefix() string
	GetKey() Key
}

const (
	Resource storeType = "resource"
	Index    storeType = "index"
)

type etcdResourceKey struct {
	prefix    string
	storeType storeType
	key       Key
}

func (e *etcdResourceKey) GetPrefix() string {
	return e.prefix
}

func (e etcdResourceKey) GetResourceType() core_model.ResourceType {
	return e.key.Typ
}

func (e etcdResourceKey) GetMesh() string {
	return e.key.Mesh
}

func (e etcdResourceKey) GetName() string {
	return e.key.Name
}

func NewEtcdResourcedKey(prefix string, typ core_model.ResourceType, mesh, name string) EtcdKey {
	return &etcdResourceKey{
		prefix:    prefix,
		storeType: Resource,
		key: Key{
			Typ:  typ,
			Mesh: mesh,
			Name: name,
		},
	}
}

func (k etcdResourceKey) String() string {
	builder := strings.Builder{}

	builder.WriteString("/")
	builder.WriteString(k.prefix)

	builder.WriteString("/")
	builder.WriteString(k.storeType.String())

	builder.WriteString(k.key.String())

	return builder.String()
}

func (k etcdResourceKey) Prefix() string {
	builder := strings.Builder{}
	builder.WriteString("/")
	if len(k.prefix) != 0 {
		builder.WriteString(k.prefix)
	} else {
		goto r
	}

	if len(k.storeType.String()) != 0 {
		builder.WriteString("/")
		builder.WriteString(k.storeType.String())
	} else {
		goto r
	}
	if len(k.key.Typ) != 0 {
		builder.WriteString("/")
		builder.WriteString(string(k.key.Typ))
	} else {
		goto r
	}

	if len(k.key.Mesh) != 0 {
		builder.WriteString("/")
		builder.WriteString(k.key.Mesh)
	} else {
		goto r
	}
	if len(k.key.Name) != 0 {
		builder.WriteString("/")
		builder.WriteString(k.key.Name)
	} else {
		goto r
	}

r:
	return builder.String()
}

func (k etcdResourceKey) GetKey() Key {
	return k.key
}

func WithEtcdKey(key string) (EtcdKey, error) {
	split := strings.Split(key, "/")

	switch split[2] {
	case Resource.String():
		if len(split) != 6 {
			return nil, errors.New("key format error ")
		}
		return &etcdResourceKey{
			prefix:    split[1],
			storeType: Resource,
			key: Key{
				Typ:  core_model.ResourceType(split[3]),
				Mesh: split[4],
				Name: split[5],
			},
		}, nil
	case Index.String():
		if len(split) != 9 {
			return nil, errors.New("key format error ")
		}
		return &etcdIndexKey{
			prefix:    split[1],
			storeType: Index,
			owner: &Key{
				Typ:  core_model.ResourceType(split[3]),
				Mesh: split[4],
				Name: split[5],
			},
			key: &Key{
				Typ:  core_model.ResourceType(split[6]),
				Mesh: split[7],
				Name: split[8],
			},
		}, nil

	default:
		return nil, errors.New("key format error ")
	}
}

type Key struct {
	Typ  core_model.ResourceType
	Mesh string
	Name string
}

func (k Key) String() string {
	builder := strings.Builder{}

	builder.WriteString("/")
	if len(k.Typ) != 0 {
		builder.WriteString(string(k.Typ))
	}

	builder.WriteString("/")
	if len(k.Mesh) != 0 {
		builder.WriteString(k.Mesh)
	}

	builder.WriteString("/")
	if len(k.Name) != 0 {
		builder.WriteString(k.Name)
	}
	return builder.String()
}

type etcdIndexKey struct {
	prefix    string
	storeType storeType
	key       *Key
	owner     *Key
}

func (e etcdIndexKey) String() string {
	builder := strings.Builder{}

	builder.WriteString("/")
	builder.WriteString(e.prefix)

	builder.WriteString("/")
	builder.WriteString(e.storeType.String())

	if e.owner != nil {
		builder.WriteString(e.owner.String())
	}

	if e.key != nil {
		builder.WriteString(e.key.String())
	}

	return builder.String()
}

func (e etcdIndexKey) Prefix() string {
	return ""
}

func (k etcdIndexKey) GetKey() Key {
	return *k.key
}

func NewEtcdIndexKey(prefix string, key *Key, owner *Key) EtcdKey {
	return &etcdIndexKey{
		prefix:    prefix,
		storeType: Index,
		key:       key,
		owner:     owner,
	}
}
