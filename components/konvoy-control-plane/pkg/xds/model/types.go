package model

import "fmt"

type ProxyId struct {
	Name      string
	Namespace string
}

func (id *ProxyId) String() string {
	return fmt.Sprintf("%s.%s", id.Name, id.Namespace)
}

type Proxy struct {
	Id       ProxyId
	Workload Workload
}

type WorkloadMeta interface {
	GetName() string
	GetNamespace() string
	GetAnnotations() map[string]string
}

type Workload struct {
	Meta      WorkloadMeta
	Version   string
	Addresses []string
	Ports     []uint32
}
