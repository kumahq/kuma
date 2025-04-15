package proto_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

type sampleStruct struct {
	Name   string `json:"name"`
	Age    uint32 `json:"age"`
	Active bool   `json:"active"`
	Nested struct {
		Role string `json:"role"`
	} `json:"nested"`
}

var _ = Describe("Proto helpers", func() {
	obj := sampleStruct{
		Name:   "Alice",
		Age:    30,
		Active: true,
	}
	obj.Nested.Role = "admin"

	expected := map[string]any{
		"name":   "Alice",
		"age":    float64(30),
		"active": true,
		"nested": map[string]any{
			"role": "admin",
		},
	}

	It("StructToMapOfAny should convert struct to map", func() {
		out := util_proto.StructToMapOfAny(obj)
		Expect(out).To(Equal(expected))
	})

	It("StructToProtoStruct should convert struct to *structpb.Struct", func() {
		st, err := util_proto.StructToProtoStruct(obj)
		Expect(err).ToNot(HaveOccurred())
		Expect(st.AsMap()).To(Equal(expected))
	})

	It("MustStructToProtoStruct should panic on invalid input", func() {
		Expect(func() {
			_ = util_proto.MustStructToProtoStruct(make(chan int)) // not serializable
		}).To(Panic())
	})

	It("MustStructToProtoStruct should return structpb.Struct", func() {
		st := util_proto.MustStructToProtoStruct(obj)
		Expect(st.AsMap()).To(Equal(expected))
	})

	It("MustFromMapOfAny should convert map to struct", func() {
		out := util_proto.MustFromMapOfAny[sampleStruct](expected)
		Expect(out).To(Equal(obj))
	})

	It("MustFromMapOfAny should convert structpb.Struct to struct", func() {
		st := util_proto.MustStructToProtoStruct(obj)
		out := util_proto.MustFromMapOfAny[sampleStruct](st)
		Expect(out).To(Equal(obj))
	})

	It("FromMapOfAny should return error on invalid JSON", func() {
		badMap := map[string]any{"invalid": make(chan int)}
		_, err := util_proto.FromMapOfAny[sampleStruct](badMap)
		Expect(err).To(HaveOccurred())
	})

	It("FromMapOfAny should convert map or structpb.Struct to struct", func() {
		st := util_proto.MustStructToProtoStruct(obj)

		out1, err1 := util_proto.FromMapOfAny[sampleStruct](expected)
		Expect(err1).ToNot(HaveOccurred())
		Expect(out1).To(Equal(obj))

		out2, err2 := util_proto.FromMapOfAny[sampleStruct](st)
		Expect(err2).ToNot(HaveOccurred())
		Expect(out2).To(Equal(obj))
	})
})
