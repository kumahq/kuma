package generator_test

import (
	"context"
	"crypto/sha256"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"time"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/xds"
	bldrs_common "github.com/kumahq/kuma/v3/pkg/envoy/builders/common"
	bldrs_core "github.com/kumahq/kuma/v3/pkg/envoy/builders/core"
	bldrs_tls "github.com/kumahq/kuma/v3/pkg/envoy/builders/tls"
	. "github.com/kumahq/kuma/v3/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/v3/pkg/test/resources/model"
	"github.com/kumahq/kuma/v3/pkg/tls"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
	xds_context "github.com/kumahq/kuma/v3/pkg/xds/context"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
	"github.com/kumahq/kuma/v3/pkg/xds/generator"
)

var _ = Describe("AdminProxyGenerator", func() {
	generator := generator.AdminProxyGenerator{}

	type testCase struct {
		dataplaneFile   string
		expected        string
		adminAddress    string
		adminSocketPath string
		readinessPort   uint32
	}

	DescribeTable("should generate envoy config",
		func(given testCase) {
			// given

			// dataplane
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := os.ReadFile(filepath.Join("testdata", "admin", given.dataplaneFile))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)

			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
						Spec: &mesh_proto.Mesh{},
					},
				},
			}

			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:       9901,
					AdminAddress:    given.adminAddress,
					AdminSocketPath: given.adminSocketPath,
					ReadinessPort:   given.readinessPort,
					IPv6Enabled:     true,
				},
				EnvoyAdminMTLSCerts: xds.ServerSideMTLSCerts{
					CaPEM: []byte("caPEM"),
					ServerPair: tls.KeyPair{
						CertPEM: []byte("certPEM"),
						KeyPEM:  []byte("keyPEM"),
					},
				},
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV3,
			}

			// when
			resources, err := generator.Generate(context.Background(), nil, ctx, proxy)

			// then
			Expect(err).ToNot(HaveOccurred())

			resp, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(resp)
			Expect(err).ToNot(HaveOccurred())

			// and output matches golden files
			Expect(actual).To(MatchGoldenYAML(filepath.Join("testdata", "admin", given.expected)))
		},
		Entry("should generate admin resources, empty admin address, readiness with TCP port 9902", testCase{
			dataplaneFile: "01.dataplane.input.yaml",
			expected:      "01.envoy-config.golden.yaml",
			adminAddress:  "",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, IPv4 loopback, readiness with TCP port 9902", testCase{
			dataplaneFile: "02.dataplane.input.yaml",
			expected:      "02.envoy-config.golden.yaml",
			adminAddress:  "127.0.0.1",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, IPv6 loopback, readiness with TCP port 9902", testCase{
			dataplaneFile: "03.dataplane.input.yaml",
			expected:      "03.envoy-config.golden.yaml",
			adminAddress:  "::1",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, unspecified IPv4, readiness with TCP port 9902", testCase{
			dataplaneFile: "04.dataplane.input.yaml",
			expected:      "04.envoy-config.golden.yaml",
			adminAddress:  "0.0.0.0",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, unspecified IPv6, readiness with TCP port 9902", testCase{
			dataplaneFile: "05.dataplane.input.yaml",
			expected:      "05.envoy-config.golden.yaml",
			adminAddress:  "::",
			readinessPort: 9902,
		}),
		Entry("should generate admin resources, Unix socket disabled, IPv6 with readiness with TCP port 9400", testCase{
			dataplaneFile: "06.dataplane.input.yaml",
			expected:      "06.envoy-config.golden.yaml",
			adminAddress:  "::1",
			readinessPort: 9400,
		}),
		Entry("should generate admin resources, admin with Unix socket", testCase{
			dataplaneFile:   "09.dataplane.input.yaml",
			expected:        "09.envoy-config.golden.yaml",
			adminAddress:    "127.0.0.1",
			adminSocketPath: "/tmp/kuma-dp/kuma-envoy-admin.sock",
			readinessPort:   9902,
		}),
	)

	DescribeTable("should return error",
		func(given testCase) {
			ctx := xds_context.Context{
				Mesh: xds_context.MeshContext{
					Resource: &core_mesh.MeshResource{
						Meta: &test_model.ResourceMeta{
							Name: "default",
						},
					},
				},
			}
			proxy := &xds.Proxy{
				Id: *xds.BuildProxyId("default", "test-admin-dpp"),
				Metadata: &xds.DataplaneMetadata{
					AdminPort:     9901,
					AdminAddress:  given.adminAddress,
					ReadinessPort: given.readinessPort,
				},
				EnvoyAdminMTLSCerts: xds.ServerSideMTLSCerts{
					CaPEM: []byte("caPEM"),
					ServerPair: tls.KeyPair{
						CertPEM: []byte("certPEM"),
						KeyPEM:  []byte("keyPEM"),
					},
				},
				Dataplane:  core_mesh.NewDataplaneResource(),
				APIVersion: envoy_common.APIV3,
				// internal addresses are set to "localhost" addresses to the "admin" listener
				// because user set x-envoy headers do not apply to this listener
				// we are settings these values here to assert they should not be generated onto Envoy config of the listener
				InternalAddresses: []xds.InternalAddress{
					{AddressPrefix: "10.0.0.0", PrefixLen: 8},
					{AddressPrefix: "127.0.0.1", PrefixLen: 32},
				},
			}

			// when
			_, err := generator.Generate(context.Background(), nil, ctx, proxy)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal(given.expected))
		},
		Entry("should return error when admin address is not allowed", testCase{
			expected:      `envoy admin cluster is not allowed to have addresses other than "", "0.0.0.0", "127.0.0.1", "::", "::1"`,
			adminAddress:  "192.168.0.1", // it's not allowed to use such address
			readinessPort: 9902,
		}),
		Entry("should return error when readiness port is 0", testCase{
			expected:      "ReadinessPort has to be in (0, 65535] range",
			adminAddress:  "127.0.0.1",
			readinessPort: 0,
		}),
	)

	Describe("identity readiness", func() {
		newProxy := func() *xds.Proxy {
			dataplane := core_mesh.NewDataplaneResource()
			bytes, err := os.ReadFile(filepath.Join("testdata", "admin", "01.dataplane.input.yaml"))
			Expect(err).ToNot(HaveOccurred())
			parseResource(bytes, dataplane)
			return &xds.Proxy{
				Metadata: &xds.DataplaneMetadata{
					AdminPort:     9901,
					AdminAddress:  "127.0.0.1",
					ReadinessPort: 9902,
					WorkDir:       "/tmp/kuma-dp",
				},
				Dataplane:  dataplane,
				APIVersion: envoy_common.APIV3,
			}
		}

		generateYAML := func(proxy *xds.Proxy) (*xds.ResourceSet, string) {
			resources, err := generator.Generate(context.Background(), nil, xds_context.Context{}, proxy)
			Expect(err).ToNot(HaveOccurred())
			response, err := resources.List().ToDeltaDiscoveryResponse()
			Expect(err).ToNot(HaveOccurred())
			actual, err := util_proto.ToYAML(response)
			Expect(err).ToNot(HaveOccurred())
			return resources, string(actual)
		}

		It("does not require identity when no MeshIdentity targets the proxy", func() {
			_, actual := generateYAML(newProxy())
			Expect(actual).To(ContainSubstring(`inlineString: '{"required":false}'`))
			Expect(actual).ToNot(ContainSubstring("name: system_identity_readiness\n"))
		})

		It("fails closed while a required identity is uninitialized", func() {
			proxy := newProxy()
			proxy.WorkloadIdentityRequired = true
			_, actual := generateYAML(proxy)
			Expect(actual).To(ContainSubstring(`inlineString: '{"required":true}'`))
			Expect(actual).ToNot(ContainSubstring("name: system_identity_readiness\n"))
		})

		It("generates an exact identity certificate sentinel", func() {
			proxy := newProxy()
			generatedAt := time.Date(2026, time.September, 8, 10, 0, 0, 0, time.UTC)
			expiresAt := generatedAt.Add(time.Hour)
			certificatePEM, err := os.ReadFile(filepath.Join("..", "..", "..", "test", "certs", "server-cert.pem"))
			Expect(err).ToNot(HaveOccurred())
			identityResources := xds.NewResourceSet().Add(&xds.Resource{
				Name: "identity",
				Resource: &envoy_tls.Secret{
					Name: "identity",
					Type: &envoy_tls.Secret_TlsCertificate{
						TlsCertificate: &envoy_tls.TlsCertificate{
							CertificateChain: &envoy_core.DataSource{
								Specifier: &envoy_core.DataSource_InlineBytes{InlineBytes: certificatePEM},
							},
						},
					},
				},
			})
			proxy.WorkloadIdentityRequired = true
			proxy.WorkloadIdentity = &xds.WorkloadIdentity{
				ManagementMode:      xds.KumaManagementMode,
				GenerationTime:      &generatedAt,
				ExpirationTime:      &expiresAt,
				AdditionalResources: identityResources,
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"identity",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			}

			resources, actual := generateYAML(proxy)
			Expect(resources.Resources(envoy_resource.SecretType)).To(BeEmpty())
			block, _ := pem.Decode(certificatePEM)
			Expect(block).ToNot(BeNil())
			hash := sha256.Sum256(block.Bytes)
			Expect(actual).To(ContainSubstring(fmt.Sprintf("%x", hash)))
			Expect(actual).To(ContainSubstring("2026-09-08T11:00:00Z"))
			Expect(actual).To(ContainSubstring("name: system_identity_readiness\n"))
			Expect(actual).To(ContainSubstring("name: identity"))
		})

		It("uses the external SDS identity and omits a control-plane fingerprint", func() {
			proxy := newProxy()
			proxy.WorkloadIdentityRequired = true
			proxy.WorkloadIdentity = &xds.WorkloadIdentity{
				ManagementMode: xds.ExternalManagementMode,
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"spiffe://example.org/workload",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			}

			_, actual := generateYAML(proxy)
			Expect(actual).To(ContainSubstring("spiffe://example.org/workload"))
			Expect(actual).To(ContainSubstring("name: system_identity_readiness\n"))
			Expect(actual).ToNot(ContainSubstring("certificateHash"))
		})

		It("supports inspected managed identities without certificate resources", func() {
			proxy := newProxy()
			proxy.WorkloadIdentityRequired = true
			proxy.WorkloadIdentity = &xds.WorkloadIdentity{
				ManagementMode: xds.KumaManagementMode,
				IdentitySourceConfigurer: func() bldrs_common.Configurer[envoy_tls.SdsSecretConfig] {
					return bldrs_tls.SdsSecretConfigSource(
						"identity",
						bldrs_core.NewConfigSource().Configure(bldrs_core.Sds()),
					)
				},
			}

			_, actual := generateYAML(proxy)
			Expect(actual).To(ContainSubstring("name: system_identity_readiness\n"))
			Expect(actual).To(ContainSubstring("name: identity"))
			Expect(actual).ToNot(ContainSubstring("certificateHash"))
		})
	})
})
