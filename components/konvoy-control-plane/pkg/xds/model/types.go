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

type Workload struct {
	Version   string
	Addresses []string
	Ports     []uint32
}
