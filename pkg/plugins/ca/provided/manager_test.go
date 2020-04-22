package provided_test

import (
	"context"

	"github.com/ghodss/yaml"
	structpb "github.com/golang/protobuf/ptypes/struct"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/ca"
	"github.com/Kong/kuma/pkg/core/datasource"
	"github.com/Kong/kuma/pkg/plugins/ca/provided"
	"github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("", func() {
	var caManager ca.CaManager

	BeforeEach(func() {
		caManager = provided.NewProvidedCaManager(datasource.NewDataSourceLoader(nil))
	})

	type testCase struct {
		configYAML string
		expected   string
	}

	DescribeTable("should throw errors",
		func(given testCase) {
			// given
			str := structpb.Struct{}
			err := proto.FromYAML([]byte(given.configYAML), &str)
			Expect(err).ToNot(HaveOccurred())

			// when
			verr := caManager.ValidateBackend(context.Background(), "default", v1alpha1.CertificateAuthorityBackend{
				Name:   "provided-1",
				Type:   "provided",
				Config: &str,
			})

			// then
			actual, err := yaml.Marshal(verr)
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(MatchYAML(given.expected))
		},
		Entry("empty config", testCase{
			configYAML: ``,
			expected: `
            violations:
            - field: cert
              message: has to be defined
            - field: key
              message: has to be defined`,
		}),
		Entry("config without data source", testCase{
			configYAML: `
            cert: {}
            key: {}`,
			expected: `
            violations:
            - field: cert
              message: 'data source has to be chosen. Available sources: secret, file, inline'
            - field: key
              message: 'data source has to be chosen. Available sources: secret, file, inline'`,
		}),
		Entry("config with empty secret", testCase{
			configYAML: `
            cert:
              secret:
            key:
              secret:`,
			expected: `
            violations:
            - field: cert.secret
              message: cannot be empty
            - field: key.secret
              message: cannot be empty`,
		}),
		Entry("config with empty secret", testCase{
			configYAML: `
            cert:
              file: '/tmp/non-existing-file'
            key:
              file: /tmp/non-existing-file`,
			expected: `
            violations:
            - field: cert
              message: 'could not load data: open /tmp/non-existing-file: no such file or directory'
            - field: key
              message: 'could not load data: open /tmp/non-existing-file: no such file or directory'`,
		}),
		Entry("config with invalid cert", testCase{
			configYAML: `
            cert:
              inline: dGVzdA==
            key:
              inline: dGVzdA==`,
			expected: `
            violations:
            - field: cert
              message: 'not a valid TLS key pair: tls: failed to find any PEM data in certificate input'`,
		}),
	)
})
