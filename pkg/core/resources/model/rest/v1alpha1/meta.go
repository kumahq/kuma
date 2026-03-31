package v1alpha1

import (
	"encoding/json"
	"time"

	"github.com/kumahq/kuma/v2/pkg/core/kri"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
)

type ResourceMeta struct {
	Type             string            `json:"type"`
	Mesh             string            `json:"mesh,omitempty"`
	Name             string            `json:"name"`
	CreationTime     time.Time         `json:"creationTime"`
	ModificationTime time.Time         `json:"modificationTime"`
	Labels           map[string]string `json:"labels,omitempty"`
}

func (r ResourceMeta) MarshalJSON() ([]byte, error) {
	type Alias ResourceMeta // prevent recursion
	out := struct {
		Alias
		KRI string `json:"kri,omitempty"`
	}{
		Alias: Alias(r),
	}

	if id, _ := kri.FromResourceMetaE(r, r.Type); !id.IsEmpty() {
		out.KRI = id.String()
	}

	return json.Marshal(&out)
}

var _ core_model.ResourceMeta = ResourceMeta{}

func (r ResourceMeta) GetName() string {
	return r.Name
}

func (r ResourceMeta) GetNameExtensions() core_model.ResourceNameExtensions {
	return core_model.ResourceNameExtensionsUnsupported
}

func (r ResourceMeta) GetVersion() string {
	return ""
}

func (r ResourceMeta) GetMesh() string {
	return r.Mesh
}

func (r ResourceMeta) GetCreationTime() time.Time {
	return r.CreationTime
}

func (r ResourceMeta) GetModificationTime() time.Time {
	return r.ModificationTime
}

func (r ResourceMeta) GetLabels() map[string]string {
	return r.Labels
}
