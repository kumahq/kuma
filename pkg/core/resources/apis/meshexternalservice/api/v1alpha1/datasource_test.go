package v1alpha1_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"sigs.k8s.io/yaml"

	datasource_api "github.com/kumahq/kuma/v2/api/common/v1alpha1/datasource"
	system_proto "github.com/kumahq/kuma/v2/api/system/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/apis/meshexternalservice/api/v1alpha1"
	core_model "github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v2/pkg/test/resources/model"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

var _ = Describe("VerificationDataSource", func() {
	DescribeTable("ToProto",
		func(given v1alpha1.VerificationDataSource, expected *system_proto.DataSource) {
			actual, err := given.ToProto()
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(matchers.MatchProto(expected))
		},
		Entry("legacy secret",
			v1alpha1.VerificationDataSource{Secret: pointer.To("my-secret")},
			&system_proto.DataSource{Type: &system_proto.DataSource_Secret{Secret: "my-secret"}},
		),
		Entry("legacy inline",
			v1alpha1.VerificationDataSource{Inline: pointer.To([]byte("test"))},
			&system_proto.DataSource{Type: &system_proto.DataSource_Inline{Inline: &wrapperspb.BytesValue{Value: []byte("test")}}},
		),
		Entry("legacy inlineString",
			v1alpha1.VerificationDataSource{InlineString: pointer.To("test")},
			&system_proto.DataSource{Type: &system_proto.DataSource_InlineString{InlineString: "test"}},
		),
		Entry("secret reference",
			v1alpha1.VerificationDataSource{
				Type:      pointer.To(datasource_api.SecureDataSourceSecretRef),
				SecretRef: &datasource_api.SecretRef{Kind: datasource_api.SecretRefType, Name: "my-secret"},
			},
			&system_proto.DataSource{Type: &system_proto.DataSource_Secret{Secret: "my-secret"}},
		),
		Entry("insecure inline",
			v1alpha1.VerificationDataSource{
				Type:           pointer.To(datasource_api.SecureDataSourceInline),
				InsecureInline: &datasource_api.Inline{Value: "test"},
			},
			&system_proto.DataSource{Type: &system_proto.DataSource_InlineString{InlineString: "test"}},
		),
	)

	DescribeTable("ToProto rejects",
		func(given v1alpha1.VerificationDataSource, expectedErr string) {
			_, err := given.ToProto()
			Expect(err).To(MatchError(expectedErr))
		},
		Entry("secret type without secretRef",
			v1alpha1.VerificationDataSource{Type: pointer.To(datasource_api.SecureDataSourceSecretRef)},
			"secretRef must be defined",
		),
		Entry("insecure inline type without value",
			v1alpha1.VerificationDataSource{Type: pointer.To(datasource_api.SecureDataSourceInline)},
			"insecureInline must be defined",
		),
		Entry("file type",
			v1alpha1.VerificationDataSource{Type: pointer.To(datasource_api.SecureDataSourceFile)},
			"datasource type: File is not supported on MeshExternalService",
		),
		Entry("env var type",
			v1alpha1.VerificationDataSource{Type: pointer.To(datasource_api.SecureDataSourceEnvVar)},
			"datasource type: EnvVar is not supported on MeshExternalService",
		),
	)

	DescribeTable("decodes the supported 3.0 SecureDataSource shapes without losing fields",
		func(given string) {
			ds := v1alpha1.VerificationDataSource{}
			Expect(core_model.FromYAML([]byte(given), &ds)).To(Succeed())
			actual, err := yaml.Marshal(ds)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given))
		},
		Entry("Secret", "{type: Secret, secretRef: {kind: Secret, name: my-secret}}"),
		Entry("InsecureInline", "{type: InsecureInline, insecureInline: {value: test}}"),
	)

	Describe("Deprecations()", func() {
		deprecations := func(spec string) []string {
			mes := v1alpha1.NewMeshExternalServiceResource()
			Expect(core_model.FromYAML([]byte(spec), &mes.Spec)).To(Succeed())
			mes.SetMeta(&test_model.ResourceMeta{Name: "external-service", Mesh: core_model.DefaultMesh})
			return mes.Deprecations()
		}

		It("should warn on every legacy data source", func() {
			Expect(deprecations(`
match: {type: HostnameGenerator, port: 443, protocol: http}
tls:
  verification:
    caCert: {inline: dGVzdA==}
    clientCert: {secret: client-cert}
    clientKey: {inlineString: key}
`)).To(Equal([]string{
				"'spec.tls.verification.caCert' uses 'secret', 'inline' or 'inlineString', which are deprecated and no longer read in 3.0. Use 'type: Secret' with 'secretRef', or 'type: InsecureInline' with 'insecureInline.value' (plain text, not base64).",
				"'spec.tls.verification.clientCert' uses 'secret', 'inline' or 'inlineString', which are deprecated and no longer read in 3.0. Use 'type: Secret' with 'secretRef', or 'type: InsecureInline' with 'insecureInline.value' (plain text, not base64).",
				"'spec.tls.verification.clientKey' uses 'secret', 'inline' or 'inlineString', which are deprecated and no longer read in 3.0. Use 'type: Secret' with 'secretRef', or 'type: InsecureInline' with 'insecureInline.value' (plain text, not base64).",
			}))
		})

		It("should not warn on the secure data source shape", func() {
			Expect(deprecations(`
match: {type: HostnameGenerator, port: 443, protocol: http}
tls:
  verification:
    caCert: {type: InsecureInline, insecureInline: {value: test}}
    clientCert: {type: Secret, secretRef: {kind: Secret, name: client-cert}}
    clientKey: {type: Secret, secretRef: {kind: Secret, name: client-key}}
`)).To(BeEmpty())
		})
	})
})
