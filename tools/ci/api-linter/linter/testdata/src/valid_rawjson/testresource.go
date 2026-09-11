package valid_rawjson

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
)

var ExtensionPoint = extensions.Point{
	ResourceType:   "TestResource",
	SchemaPath:     []string{"spec", "extension"},
	Discriminator:  "type",
	ConfigProperty: "config",
}

// TestResource
type TestResource struct {
	Extension *Extension `json:"extension,omitempty"` // OK
	Logging   *Logging   `json:"logging,omitempty"`   // OK
}

type Extension struct {
	// +kuma:discriminator
	Type string `json:"type"` // OK

	Config *apiextensionsv1.JSON `json:"config,omitempty"` // OK: the point above describes it
}

type Logging struct {
	// +kuma:opaque-payload // the user's own value, nothing to describe
	Body *apiextensionsv1.JSON `json:"body,omitempty"` // OK
}
