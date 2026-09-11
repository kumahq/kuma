package generator

import (
	"github.com/invopop/jsonschema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("applyFieldMarkers", func() {
	It("should promote +required to the parent schema and strip every marker", func() {
		// given
		properties := jsonschema.NewProperties()
		properties.Set("kind", &jsonschema.Schema{
			Type:        "string",
			Description: "Type of the backend\n\n\t+required",
		})
		properties.Set("name", &jsonschema.Schema{
			Type:        "string",
			Description: "Name of the backend.\n\n\t+optional",
		})
		schema := &jsonschema.Schema{Type: "object", Properties: properties}

		// when
		applyFieldMarkers(schema)

		// then
		Expect(schema.Required).To(Equal([]string{"kind"}))
		Expect(properties.Value("kind").Description).To(Equal("Type of the backend"))
		Expect(properties.Value("name").Description).To(Equal("Name of the backend."))
	})

	It("should descend into nested objects and arrays", func() {
		// given
		itemProperties := jsonschema.NewProperties()
		itemProperties.Set("address", &jsonschema.Schema{
			Type:        "string",
			Description: "IP of the outbound\n\n\t+required",
		})
		outboundProperties := jsonschema.NewProperties()
		outboundProperties.Set("outbound", &jsonschema.Schema{
			Type:  "array",
			Items: &jsonschema.Schema{Type: "object", Properties: itemProperties},
		})
		schema := &jsonschema.Schema{Type: "object", Properties: outboundProperties}

		// when
		applyFieldMarkers(schema)

		// then
		Expect(schema.Required).To(BeEmpty())
		Expect(outboundProperties.Value("outbound").Items.Required).To(Equal([]string{"address"}))
		Expect(itemProperties.Value("address").Description).To(Equal("IP of the outbound"))
	})

	It("should be idempotent", func() {
		// given
		properties := jsonschema.NewProperties()
		properties.Set("kind", &jsonschema.Schema{
			Type:        "string",
			Description: "Type of the backend\n\n\t+required",
		})
		schema := &jsonschema.Schema{Type: "object", Properties: properties}

		// when
		applyFieldMarkers(schema)
		applyFieldMarkers(schema)

		// then
		Expect(schema.Required).To(Equal([]string{"kind"}))
		Expect(properties.Value("kind").Description).To(Equal("Type of the backend"))
	})

	It("should not repeat a field the reflector already made required", func() {
		// given
		properties := jsonschema.NewProperties()
		properties.Set("kind", &jsonschema.Schema{
			Type:        "string",
			Description: "Type of the backend\n\n\t+required",
		})
		schema := &jsonschema.Schema{Type: "object", Properties: properties, Required: []string{"kind"}}

		// when
		applyFieldMarkers(schema)

		// then
		Expect(schema.Required).To(Equal([]string{"kind"}))
	})
})
