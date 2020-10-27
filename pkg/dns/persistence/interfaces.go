package persistence

type Reader interface {
	Get() (VIPList, error)
}

type MeshedWriter interface {
	Reader
	GetByMesh(mesh string) (VIPList, error)
	Set(mesh string, vips VIPList) error
}

type GlobalWriter interface {
	Reader
	Set(vips VIPList) error
}

type VIPList map[string]string

func (vips VIPList) Append(other VIPList) {
	for k, v := range other {
		vips[k] = v
	}
}
