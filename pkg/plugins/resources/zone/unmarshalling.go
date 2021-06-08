package zone

import (
	"encoding/json"
	"net/url"
	"time"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
)

type zoneMeta struct {
	Name             string
	Mesh             string
	Version          string
	CreationTime     time.Time
	ModificationTime time.Time
}

func (m zoneMeta) GetName() string {
	return m.Name
}

func (m zoneMeta) GetNameExtensions() model.ResourceNameExtensions {
	return model.ResourceNameExtensionsUnsupported
}

func (m zoneMeta) GetMesh() string {
	return m.Mesh
}

func (m zoneMeta) GetVersion() string {
	return m.Version
}

func (m zoneMeta) GetCreationTime() time.Time {
	return m.CreationTime
}

func (m zoneMeta) GetModificationTime() time.Time {
	return m.ModificationTime
}

func Unmarshal(b []byte, res model.Resource) error {
	restResource := rest.Resource{
		Spec: res.GetSpec(),
	}
	if err := json.Unmarshal(b, &restResource); err != nil {
		return err
	}
	res.SetMeta(zoneMeta{
		Name:             restResource.Meta.Name,
		Mesh:             restResource.Meta.Mesh,
		Version:          "",
		CreationTime:     restResource.Meta.CreationTime,
		ModificationTime: restResource.Meta.ModificationTime,
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
		r.SetMeta(&zoneMeta{
			Name:             ri.Meta.Name,
			Mesh:             ri.Meta.Mesh,
			Version:          "",
			CreationTime:     ri.Meta.CreationTime,
			ModificationTime: ri.Meta.ModificationTime,
		})
		_ = rs.AddItem(r)
	}
	if rsr.Next != nil {
		uri, err := url.ParseRequestURI(*rsr.Next)
		if err != nil {
			return errors.Wrap(err, "invalid next URL from the server")
		}
		offset := uri.Query().Get("offset")
		// we do not preserve here the size of the page, but since it is used in kumactl
		// user will rerun command with the page size of his choice
		if offset != "" {
			rs.GetPagination().SetNextOffset(offset)
		}
	}
	rs.GetPagination().SetTotal(rsr.ResourceList.Total)
	return nil
}
