package main

import (
	"github.com/invopop/jsonschema"
	"sigs.k8s.io/yaml"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
)

func main() {

	reflector := jsonschema.Reflector{
		ExpandedStruct:            true,
		DoNotReference:            true,
		AllowAdditionalProperties: true,
	}
	err := reflector.AddGoComments("github.com/kumahq/kuma/", "api/mesh/v1alpha1/")
	if err != nil {
		return
	}
	schema := reflector.Reflect(&v1alpha1.Mesh{})
	//out, _ := json.MarshalIndent(schema, "", "  ")
	out, _ := yaml.Marshal(schema)

	println(string(out))
}
