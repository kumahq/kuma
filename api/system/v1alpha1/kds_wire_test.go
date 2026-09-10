package v1alpha1_test

import (
	"encoding/hex"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/anypb"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
)

// The type URLs and bytes below are what a 2.14 control plane put on the KDS
// wire for these two specs. A zone of any released version reads the Any by its
// type URL and then parses protobuf, and fails the whole DeltaDiscoveryResponse
// when either does not match, so a global that changes them cannot sync to that
// zone at all.
var _ = Describe("KDS wire compatibility", func() {
	DescribeTable("writes the secret bytes 2.14 wrote",
		func(spec *system_proto.Secret, encoded string) {
			any, err := core_model.ToAny(spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.system.v1alpha1.Secret"))
			Expect(hex.EncodeToString(any.GetValue())).To(Equal(encoded))

			read := &system_proto.Secret{}
			Expect(core_model.FromAny(any, read)).To(Succeed())
			Expect(read.GetData().GetValue()).To(Equal(spec.GetData().GetValue()))
		},
		Entry("a value", &system_proto.Secret{Data: system_proto.Bytes([]byte("hunter2"))}, "0a090a0768756e74657232"),
		Entry("bytes that are not text", &system_proto.Secret{Data: system_proto.Bytes([]byte{0x00, 0xff, 0x10})}, "0a050a0300ff10"),
		Entry("set to no bytes", &system_proto.Secret{Data: system_proto.Bytes([]byte{})}, "0a00"),
		Entry("never set", &system_proto.Secret{}, ""),
	)

	DescribeTable("writes the config bytes 2.14 wrote",
		func(spec *system_proto.Config, encoded string) {
			any, err := core_model.ToAny(spec)
			Expect(err).ToNot(HaveOccurred())
			Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.system.v1alpha1.Config"))
			Expect(hex.EncodeToString(any.GetValue())).To(Equal(encoded))

			read := &system_proto.Config{}
			Expect(core_model.FromAny(any, read)).To(Succeed())
			Expect(read.GetConfig()).To(Equal(spec.GetConfig()))
		},
		Entry("a json payload", &system_proto.Config{Config: `{"a":1}`}, "0a077b2261223a317d"),
		Entry("never set", &system_proto.Config{}, ""),
	)

	It("keeps a secret that was set to no bytes distinct from one never set", func() {
		set, err := core_model.ToAny(&system_proto.Secret{Data: system_proto.Bytes([]byte{})})
		Expect(err).ToNot(HaveOccurred())

		read := &system_proto.Secret{}
		Expect(core_model.FromAny(set, read)).To(Succeed())
		Expect(read.Data).ToNot(BeNil())

		unset, err := core_model.ToAny(&system_proto.Secret{})
		Expect(err).ToNot(HaveOccurred())
		Expect(core_model.FromAny(unset, read)).To(Succeed())
		Expect(read.Data).To(BeNil())
	})

	It("tolerates a nil spec the way the protobuf marshaller did", func() {
		any, err := core_model.ToAny((*system_proto.Secret)(nil))
		Expect(err).ToNot(HaveOccurred())
		Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.system.v1alpha1.Secret"))
		Expect(any.GetValue()).To(BeEmpty())

		any, err = core_model.ToAny((*system_proto.Config)(nil))
		Expect(err).ToNot(HaveOccurred())
		Expect(any.GetTypeUrl()).To(Equal("type.googleapis.com/kuma.system.v1alpha1.Config"))
		Expect(any.GetValue()).To(BeEmpty())
	})

	DescribeTable("reads the json a control plane between the rewrite and this fix sent",
		func(json string, expected []byte) {
			read := &system_proto.Secret{}
			Expect(core_model.FromAny(&anypb.Any{Value: []byte(json)}, read)).To(Succeed())
			Expect(read.GetData().GetValue()).To(Equal(expected))
		},
		Entry("a value", `{"data":"aHVudGVyMg=="}`, []byte("hunter2")),
		Entry("never set", `{}`, []byte(nil)),
	)
})
