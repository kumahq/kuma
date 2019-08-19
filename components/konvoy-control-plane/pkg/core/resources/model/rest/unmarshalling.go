package rest

import (
	"encoding/json"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/model"
)

type RemoteMeta struct {
	Namespace string
	Name      string
	Mesh      string
	Version   string
}

func (m RemoteMeta) GetName() string {
	return m.Name
}
func (m RemoteMeta) GetNamespace() string {
	return m.Namespace
}
func (m RemoteMeta) GetMesh() string {
	return m.Mesh
}
func (m RemoteMeta) GetVersion() string {
	return m.Version
}

func Unmarshal(b []byte, res model.Resource) error {
	restResource := Resource{
		Spec: res.GetSpec(),
	}
	if err := json.Unmarshal(b, &restResource); err != nil {
		return err
	}
	res.SetMeta(RemoteMeta{
		Namespace: "",
		Name:      restResource.Meta.Name,
		Mesh:      restResource.Meta.Mesh,
		Version:   "",
	})
	return nil
}

func UnmarshalList(b []byte, rs model.ResourceList) error {
	rsr := &ResourceListReceiver{
		NewResource: func() model.Resource {
			return rs.NewItem()
		},
	}
	if err := json.Unmarshal(b, rsr); err != nil {
		return err
	}
	for _, ri := range rsr.ResourceList.Items {
		r := rs.NewItem()
		r.SetSpec(ri.Spec)
		r.SetMeta(&RemoteMeta{
			Namespace: "",
			Name:      ri.Meta.Name,
			Mesh:      ri.Meta.Mesh,
			Version:   "",
		})
		_ = rs.AddItem(r)
	}
	return nil
}
