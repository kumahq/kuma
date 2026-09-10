package v1alpha1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// The bytes below are what a 2.14 control plane wrote. Secrets and configs hold operator
// entered data, so a read failure loses something a resync cannot rebuild, and a control
// plane that cannot parse a secret cannot serve identity.
var _ = Describe("System storage compatibility", func() {
	DescribeTable("round trips the secret bytes 2.14 wrote",
		func(stored string, expected []byte, set bool) {
			secret := &system_proto.Secret{}
			Expect(json.Unmarshal([]byte(stored), secret)).To(Succeed())
			Expect(secret.GetData().GetValue()).To(Equal(expected))
			Expect(secret.Data != nil).To(Equal(set))

			rewritten, err := json.Marshal(secret)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rewritten)).To(Equal(stored))
		},
		Entry("a value", `{"data":"aHVudGVyMg=="}`, []byte("hunter2"), true),
		Entry("bytes that are not text", `{"data":"AP8Q"}`, []byte{0x00, 0xff, 0x10}, true),
		Entry("set to no bytes", `{"data":""}`, []byte{}, true),
		Entry("never set", `{}`, []byte(nil), false),
	)

	It("keeps a secret that was set to no bytes distinct from one never set", func() {
		empty, err := json.Marshal(&system_proto.Secret{Data: system_proto.Bytes([]byte{})})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(empty)).To(Equal(`{"data":""}`))

		unset, err := json.Marshal(&system_proto.Secret{})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(unset)).To(Equal(`{}`))
	})

	DescribeTable("round trips the config bytes 2.14 wrote",
		func(stored string, expected string) {
			config := &system_proto.Config{}
			Expect(json.Unmarshal([]byte(stored), config)).To(Succeed())
			Expect(config.GetConfig()).To(Equal(expected))

			rewritten, err := json.Marshal(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rewritten)).To(Equal(stored))
		},
		Entry("a json payload", `{"config":"{\"a\":1}"}`, `{"a":1}`),
		Entry("a plain string", `{"config":"plain"}`, "plain"),
		Entry("never set", `{}`, ""),
	)

	It("writes the chosen data source as a top level key, the way the oneof did", func() {
		Expect(json.Marshal(&system_proto.DataSource{Secret: pointer.To("my-secret")})).
			To(BeEquivalentTo(`{"secret":"my-secret"}`))
		Expect(json.Marshal(&system_proto.DataSource{Inline: system_proto.Bytes([]byte("abc"))})).
			To(BeEquivalentTo(`{"inline":"YWJj"}`))
		Expect(json.Marshal(&system_proto.DataSource{InlineString: pointer.To("abc")})).
			To(BeEquivalentTo(`{"inlineString":"abc"}`))
	})

	It("reports which data source alternative was chosen", func() {
		Expect((&system_proto.DataSource{}).IsSet()).To(BeFalse())
		Expect((*system_proto.DataSource)(nil).IsSet()).To(BeFalse())
		Expect((&system_proto.DataSource{Secret: pointer.To("s")}).HasSecret()).To(BeTrue())
		Expect((&system_proto.DataSource{File: pointer.To("/f")}).HasFile()).To(BeTrue())
	})

	It("masks an inline data source without disclosing its bytes", func() {
		masked := (&system_proto.DataSource{Inline: system_proto.Bytes([]byte("private"))}).MaskInlineDatasource()
		Expect(masked.GetInline().GetValue()).To(Equal([]byte("***")))

		masked = (&system_proto.DataSource{InlineString: pointer.To("private")}).MaskInlineDatasource()
		Expect(masked.GetInlineString()).To(Equal("***"))

		Expect((&system_proto.DataSource{Secret: pointer.To("s")}).MaskInlineDatasource()).To(BeNil())
	})
})
