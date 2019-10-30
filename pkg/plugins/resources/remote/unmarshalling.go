package remote

import (
	"encoding/json"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
)

type remoteMeta struct {
	Namespace string
	Name      string
	Mesh      string
	Version   string
}

func (m remoteMeta) GetName() string {
	return m.Name
}
func (m remoteMeta) GetNamespace() string {
	return m.Namespace
}
func (m remoteMeta) GetMesh() string {
	return m.Mesh
}
func (m remoteMeta) GetVersion() string {
	return m.Version
}

func Unmarshal(b []byte, res model.Resource) error {
	restResource := rest.Resource{
		Spec: res.GetSpec(),
	}
	if err := json.Unmarshal(b, &restResource); err != nil {
		return err
	}
	res.SetMeta(remoteMeta{
		Namespace: "",
		Name:      restResource.Meta.Name,
		Mesh:      restResource.Meta.Mesh,
		Version:   "",
	})
	return nil
}

func UnmarshalList(b []byte, rs model.ResourceList) error {
	rsr := &rest.ResourceListReceiver{
		NewResource: rs.NewItem,
	}
	if err := json.Unmarshal(b, rsr); err != nil {
		return err
	}
	for _, ri := range rsr.ResourceList.Items {
		r := rs.NewItem()
		if err := r.SetSpec(ri.Spec); err != nil {
			return err
		}
		r.SetMeta(&remoteMeta{
			Namespace: "",
			Name:      ri.Meta.Name,
			Mesh:      ri.Meta.Mesh,
			Version:   "",
		})
		_ = rs.AddItem(r)
	}
	return nil
}
