package invalid_rawjson

import (
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

    "github.com/kumahq/kuma/v3/pkg/core/resources/extensions"
)

// The point describes spec.extension, so it says nothing about the two fields
// below it.
var ExtensionPoint = extensions.Point{
    ResourceType:   "TestResource",
    SchemaPath:     []string{"spec", "extension"},
    Discriminator:  "type",
    ConfigProperty: "config",
}

// TestResource
type TestResource struct {
    Provider *Provider `json:"provider,omitempty"` // OK
    Items    []Item    `json:"items"`              // OK
}

type Provider struct {
    // +kuma:discriminator
    Type string `json:"type"` // OK

    Config *apiextensionsv1.JSON `json:"config,omitempty"` // want `raw JSON field TestResource.Provider.Config is undocumented: declare an extensions.Point with SchemaPath \[\]string\{"spec", "provider"\} and ConfigProperty "config", or mark the field '\+kuma:opaque-payload'`
}

type Item struct {
    Config *apiextensionsv1.JSON `json:"config,omitempty"` // want `raw JSON field TestResource.Items\[\].Config cannot be reached by an extension point, so it must be marked '\+kuma:opaque-payload'`
}
