package extensions

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type fakeConfig struct {
	Field string `json:"field,omitempty"`
}

type otherConfig struct {
	Field string `json:"field,omitempty"`
}

var _ = Describe("Register", func() {
	point := Point{
		ResourceType:  "FakeResource",
		SchemaPath:    []string{"spec", "extension"},
		Discriminator: "type",
	}

	BeforeEach(func() {
		mtx.Lock()
		registered = map[string]Extension{}
		mtx.Unlock()
	})

	It("should return registrations ordered by resource type then value", func() {
		Register(Extension{Point: point, Value: "zeta", Config: &fakeConfig{}})
		Register(Extension{Point: point, Value: "alpha", Config: &otherConfig{}})
		other := Point{ResourceType: "AnotherResource", SchemaPath: []string{"spec", "extension"}, Discriminator: "name"}
		Register(Extension{Point: other, Value: "omega", Config: &fakeConfig{}})

		values := []string{}
		for _, ext := range Registered() {
			values = append(values, string(ext.Point.ResourceType)+"/"+ext.Value)
		}
		Expect(values).To(Equal([]string{"AnotherResource/omega", "FakeResource/alpha", "FakeResource/zeta"}))
	})

	It("should expose the struct type behind the config pointer", func() {
		Register(Extension{Point: point, Value: "fake", Config: &fakeConfig{}})

		Expect(Registered()[0].ConfigType().Name()).To(Equal("fakeConfig"))
	})

	It("should reject a second registration of the same discriminator value", func() {
		Register(Extension{Point: point, Value: "fake", Config: &fakeConfig{}})

		Expect(func() {
			Register(Extension{Point: point, Value: "fake", Config: &otherConfig{}})
		}).To(PanicWith(MatchError(ContainSubstring(`extension "fake" is already registered on FakeResource`))))
	})

	It("should reject a resource declaring two different extension points", func() {
		Register(Extension{Point: point, Value: "fake", Config: &fakeConfig{}})

		moved := point
		moved.SchemaPath = []string{"spec", "provider", "extension"}
		Expect(func() {
			Register(Extension{Point: moved, Value: "other", Config: &otherConfig{}})
		}).To(PanicWith(MatchError(ContainSubstring("declares a different extension point on FakeResource"))))
	})

	DescribeTable("should reject a registration that cannot produce a schema",
		func(ext Extension, msg string) {
			Expect(func() { Register(ext) }).To(PanicWith(MatchError(ContainSubstring(msg))))
		},
		Entry("no resource type",
			Extension{Point: Point{SchemaPath: []string{"spec"}, Discriminator: "type"}, Value: "fake", Config: &fakeConfig{}},
			"point resource type must not be empty"),
		Entry("no schema path",
			Extension{Point: Point{ResourceType: "FakeResource", Discriminator: "type"}, Value: "fake", Config: &fakeConfig{}},
			"point schema path must not be empty"),
		Entry("no discriminator",
			Extension{Point: Point{ResourceType: "FakeResource", SchemaPath: []string{"spec"}}, Value: "fake", Config: &fakeConfig{}},
			"point discriminator must not be empty"),
		Entry("no value",
			Extension{Point: point, Config: &fakeConfig{}},
			"value must not be empty"),
		Entry("nil config",
			Extension{Point: point, Value: "fake"},
			"config must not be nil"),
		Entry("config is not a pointer",
			Extension{Point: point, Value: "fake", Config: fakeConfig{}},
			"config must be a pointer to a struct"),
		Entry("config is a pointer to a map",
			Extension{Point: point, Value: "fake", Config: &map[string]string{}},
			"config must be a pointer to a struct"),
		Entry("config is an anonymous struct",
			Extension{Point: point, Value: "fake", Config: &struct{ Field string }{}},
			"config must be a named struct type"),
	)
})
