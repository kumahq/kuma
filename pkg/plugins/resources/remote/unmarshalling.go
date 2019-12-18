package remote

import (
	"encoding/json"
	"github.com/Kong/kuma/pkg/core/resources/model"
	"github.com/Kong/kuma/pkg/core/resources/model/rest"
	"time"
)

type remoteMeta struct {
	Name             string
	Mesh             string
	Version          string
	CreationTime     time.Time
	ModificationTime time.Time
}

func (m remoteMeta) GetName() string {
	return m.Name
}
func (m remoteMeta) GetMesh() string {
	return m.Mesh
}
func (m remoteMeta) GetVersion() string {
	return m.Version
}
func (m remoteMeta) GetCreationTime() time.Time {
	return m.CreationTime
}
func (m remoteMeta) GetModificationTime() time.Time {
	return m.ModificationTime
}

func Unmarshal(b []byte, res model.Resource) error {
	restResource := rest.Resource{
		Spec: res.GetSpec(),
	}
	if err := json.Unmarshal(b, &restResource); err != nil {
		return err
	}
	res.SetMeta(remoteMeta{
		Name:    restResource.Meta.Name,
		Mesh:    restResource.Meta.Mesh,
		Version: "",
		// todo(jakubdyszkiewicz) creation and modification time is not set because it's not exposed in API yet
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
			Name:    ri.Meta.Name,
			Mesh:    ri.Meta.Mesh,
			Version: "",
			// todo(jakubdyszkiewicz) creation and modification time is not set because it's not exposed in API yet
		})
		_ = rs.AddItem(r)
	}
	return nil
}
