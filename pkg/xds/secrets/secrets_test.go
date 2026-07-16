package secrets_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"

	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core"
	core_ca "github.com/kumahq/kuma/v3/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/v3/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/resources/store"
	"github.com/kumahq/kuma/v3/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/v3/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/v3/pkg/core/secrets/store"
	core_metrics "github.com/kumahq/kuma/v3/pkg/metrics"
	ca_builtin "github.com/kumahq/kuma/v3/pkg/plugins/ca/builtin"
	"github.com/kumahq/kuma/v3/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/v3/pkg/test/metrics"
	"github.com/kumahq/kuma/v3/pkg/test/resources/model"
	. "github.com/kumahq/kuma/v3/pkg/xds/secrets"
)

var _ = Describe("Secrets", Ordered, func() {
	var secrets Secrets
	var metrics core_metrics.Metrics
	var now time.Time

	newMesh := func(name string) *core_mesh.MeshResource {
		return &core_mesh.MeshResource{
			Meta: &model.ResourceMeta{
				Name: name,
			},
			Spec: &mesh_proto.Mesh{
				Mtls: &mesh_proto.Mesh_Mtls{
					EnabledBackend: "ca-1",
					Backends: []*mesh_proto.CertificateAuthorityBackend{
						{
							Name: "ca-1",
							Type: "builtin",
							DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
								Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
									Expiration: "1h",
								},
							},
						},
						{
							Name: "ca-2",
							Type: "builtin",
							DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
								Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
									Expiration: "1h",
								},
							},
						},
					},
				},
			},
		}
	}

	newDataplane := func() *core_mesh.DataplaneResource {
		return &core_mesh.DataplaneResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "dp1",
			},
			Spec: &mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "192.168.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"kuma.io/service": "web",
							},
						},
					},
				},
			},
		}
	}

	newZoneEgress := func() *core_mesh.ZoneEgressResource {
		return &core_mesh.ZoneEgressResource{
			Meta: &model.ResourceMeta{
				Mesh: "default",
				Name: "ze-1",
			},
			Spec: &mesh_proto.ZoneEgress{
				Networking: &mesh_proto.ZoneEgress_Networking{
					Address: "192.168.0.1",
					Port:    10002,
				},
			},
		}
	}

	BeforeAll(func() {
		// since we actually create a mesh, and it goes through validation we have a default limit of 1
		core_mesh.AllowedMTLSBackends = 2
	})

	AfterAll(func() {
		core_mesh.AllowedMTLSBackends = 1
	})

	BeforeEach(func() {
		resStore := memory.NewStore()
		rm := manager.NewResourceManager(resStore)
		secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None(), nil, false)
		builtinCaManager := ca_builtin.NewBuiltinCaManager(secretManager)
		caManagers := core_ca.Managers{
			"builtin": builtinCaManager,
		}
		mesh := newMesh("default")
		err := rm.Create(context.Background(), mesh, store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = builtinCaManager.EnsureBackends(context.Background(), mesh, mesh.Spec.Mtls.Backends)
		Expect(err).ToNot(HaveOccurred())

		mesh1 := newMesh("mesh-1")
		err = rm.Create(context.Background(), mesh1, store.CreateByKey("mesh-1", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = builtinCaManager.EnsureBackends(context.Background(), mesh1, mesh1.Spec.Mtls.Backends)
		Expect(err).ToNot(HaveOccurred())

		mesh2 := newMesh("mesh-2")
		err = rm.Create(context.Background(), mesh2, store.CreateByKey("mesh-2", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
		err = builtinCaManager.EnsureBackends(context.Background(), mesh2, mesh2.Spec.Mtls.Backends)
		Expect(err).ToNot(HaveOccurred())

		m, err := core_metrics.NewMetrics("local")
		Expect(err).ToNot(HaveOccurred())
		metrics = m

		caProvider, err := NewCaProvider(caManagers, metrics)
		Expect(err).ToNot(HaveOccurred())
		identityProvider, err := NewIdentityProvider(caManagers, metrics)
		Expect(err).ToNot(HaveOccurred())

		secrets, err = NewSecrets(caProvider, identityProvider, metrics)
		Expect(err).ToNot(HaveOccurred())

		now = time.Now()
		core.Now = func() time.Time {
			return now
		}
	})

	Context("dataplane proxy", func() {
		It("should generate cert and emit statistic and info", func() {
			// when
			identity, ca, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

			// then certs are generated
			Expect(err).ToNot(HaveOccurred())
			Expect(identity.PemCerts).ToNot(BeEmpty())
			Expect(identity.PemKey).ToNot(BeEmpty())
			Expect(ca).To(HaveLen(1))
			Expect(ca["default"].PemCerts).ToNot(BeEmpty())

			// and info is stored
			info := secrets.Info(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(newDataplane().Meta))
			Expect(info.Generation).To(Equal(now))
			Expect(info.Expiration.Unix()).To(Equal(now.Add(1 * time.Hour).Unix()))
			Expect(info.OwnMesh.MTLS.EnabledBackend).To(Equal("ca-1"))
			Expect(info.Tags).To(Equal(mesh_proto.MultiValueTagSet{
				"kuma.io/service": map[string]bool{
					"web": true,
				},
			}))

			// and metric is published
			Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
		})

		It("should not regenerate certs if nothing has changed", func() {
			// given
			identity, cas, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)
			Expect(err).ToNot(HaveOccurred())

			// when
			newIdentity, newCa, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(identity).To(Equal(newIdentity))
			Expect(cas).To(HaveLen(1))
			Expect(newCa).To(HaveLen(1))
			Expect(cas["default"]).To(Equal(newCa["default"]))
			Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
		})

		Context("should regenerate certificate", func() {
			BeforeEach(func() {
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			It("when cross-mesh was added and previous removed", func() {
				// given
				defaultMesh := newMesh("default")

				// when
				_, caSecrets, err := secrets.GetForDataPlane(context.Background(), newDataplane(), defaultMesh, []*core_mesh.MeshResource{newMesh("mesh-1")})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
				Expect(caSecrets["default"]).ToNot(BeNil())
				Expect(caSecrets["mesh-1"]).ToNot(BeNil())

				// when
				_, caSecrets, err = secrets.GetForDataPlane(context.Background(), newDataplane(), defaultMesh, []*core_mesh.MeshResource{newMesh("mesh-2")})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
				Expect(caSecrets["default"]).ToNot(BeNil())
				Expect(caSecrets["mesh-1"]).To(BeNil())
				Expect(caSecrets["mesh-2"]).ToNot(BeNil())
			})

			It("when mTLS settings has changed", func() {
				// given
				mesh := newMesh("default")
				mesh.Spec.Mtls.EnabledBackend = "ca-2"

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), mesh, nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-2").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
			})

			It("when dp tags has changed", func() {
				// given
				dataplane := newDataplane()
				dataplane.Spec.Networking.Inbound[0].Tags["kuma.io/service"] = "web2"

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), dataplane, newMesh("default"), nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert is expiring", func() {
				// given
				now = now.Add(48*time.Minute + 1*time.Millisecond) // 4/5 of 60 minutes

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert was cleaned up", func() {
				// given
				secrets.Cleanup(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(newDataplane().Meta))

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(2)))
			})
		})

		It("should cleanup certs", func() {
			// given
			_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)
			Expect(err).ToNot(HaveOccurred())

			// when
			secrets.Cleanup(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(newDataplane().Meta))

			// then
			Expect(secrets.Info(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(newDataplane().Meta))).To(BeNil())
		})
	})

	Context("zone egress", func() {
		It("should generate cert and emit statistic and info", func() {
			// when
			identity, ca, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

			// then certs are generated
			Expect(err).ToNot(HaveOccurred())
			Expect(identity.PemCerts).ToNot(BeEmpty())
			Expect(identity.PemKey).ToNot(BeEmpty())
			Expect(ca.PemCerts).ToNot(BeEmpty())

			// and info is stored
			info := secrets.Info(mesh_proto.EgressProxyType, core_model.MetaToResourceKey(newZoneEgress().Meta))
			Expect(info.Generation).To(Equal(now))
			Expect(info.Expiration.Unix()).To(Equal(now.Add(1 * time.Hour).Unix()))
			Expect(info.OwnMesh.MTLS.EnabledBackend).To(Equal("ca-1"))
			Expect(info.Tags).To(Equal(mesh_proto.MultiValueTagSet{
				"kuma.io/service": map[string]bool{
					mesh_proto.ZoneEgressServiceName: true,
				},
			}))

			// and metric is published
			Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
		})

		It("should not regenerate certs if nothing has changed", func() {
			// given
			identity, ca, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			newIdentity, newCa, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

			// then
			Expect(err).ToNot(HaveOccurred())
			Expect(identity).To(Equal(newIdentity))
			Expect(ca).To(Equal(newCa))
			Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
		})

		Context("should regenerate certificate", func() {
			BeforeEach(func() {
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))
				Expect(err).ToNot(HaveOccurred())
			})

			It("when mTLS settings has changed", func() {
				// given
				mesh := newMesh("default")
				mesh.Spec.Mtls.EnabledBackend = "ca-2"

				// when
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), mesh)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-2").GetHistogram().GetSampleCount()).To(Equal(uint64(1)))
			})

			It("when cert is expiring", func() {
				// given
				now = now.Add(48*time.Minute + 1*time.Millisecond) // 4/5 of 60 minutes

				// when
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert was cleaned up", func() {
				// given
				secrets.Cleanup(mesh_proto.EgressProxyType, core_model.MetaToResourceKey(newZoneEgress().Meta))

				// when
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetHistogram().GetSampleCount()).To(Equal(uint64(2)))
			})
		})

		It("should cleanup certs", func() {
			// given
			_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))
			Expect(err).ToNot(HaveOccurred())

			// when
			secrets.Cleanup(mesh_proto.EgressProxyType, core_model.MetaToResourceKey(newZoneEgress().Meta))

			// then
			Expect(secrets.Info(mesh_proto.EgressProxyType, core_model.MetaToResourceKey(newZoneEgress().Meta))).To(BeNil())
		})
	})

	Context("certificate generation backoff", func() {
		var calls int
		var failingSecrets Secrets
		var failMetrics core_metrics.Metrics

		failingMesh := func() *core_mesh.MeshResource {
			return &core_mesh.MeshResource{
				Meta: &model.ResourceMeta{Name: "default"},
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: "ca-1",
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{
								Name: "ca-1",
								Type: "failing",
								DpCert: &mesh_proto.CertificateAuthorityBackend_DpCert{
									Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{
										Expiration: "1h",
									},
								},
							},
						},
					},
				},
			}
		}

		BeforeEach(func() {
			calls = 0
			caManagers := core_ca.Managers{
				"failing": &failingCaManager{calls: &calls},
			}

			m, err := core_metrics.NewMetrics("local")
			Expect(err).ToNot(HaveOccurred())
			failMetrics = m

			caProvider, err := NewCaProvider(caManagers, m)
			Expect(err).ToNot(HaveOccurred())
			identityProvider, err := NewIdentityProvider(caManagers, m)
			Expect(err).ToNot(HaveOccurred())
			failingSecrets, err = NewSecrets(caProvider, identityProvider, m)
			Expect(err).ToNot(HaveOccurred())

			CertGenerationBackoffBase = 5 * time.Second
			CertGenerationBackoffMax = 5 * time.Minute
		})

		It("should not call the CA on every tick after a failure", func() {
			// when the first attempt fails
			_, _, err := failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)

			// then it starts a backoff and records a failure
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(1))
			Expect(test_metrics.FindMetric(failMetrics, "cert_generation_failure").GetCounter().GetValue()).To(Equal(1.0))

			// when subsequent ticks happen within the backoff window
			for range 5 {
				_, _, err = failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)
				Expect(err).To(HaveOccurred())
			}

			// then the CA is not called again
			Expect(calls).To(Equal(1))

			// when the backoff window passes
			now = now.Add(CertGenerationBackoffBase + time.Second)
			_, _, err = failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)

			// then the CA is retried
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(2))
		})
	})
})

// failingCaManager is a ca.Manager whose GenerateDataplaneCert always fails,
// used to exercise the certificate generation backoff.
type failingCaManager struct {
	calls *int
}

var _ core_ca.Manager = &failingCaManager{}

func (f *failingCaManager) GenerateDataplaneCert(context.Context, string, *mesh_proto.CertificateAuthorityBackend, mesh_proto.MultiValueTagSet) (core_ca.KeyPair, error) {
	*f.calls++
	return core_ca.KeyPair{}, errors.New("failing CA")
}

func (f *failingCaManager) ValidateBackend(context.Context, string, *mesh_proto.CertificateAuthorityBackend) error {
	return nil
}

func (f *failingCaManager) EnsureBackends(context.Context, core_model.Resource, []*mesh_proto.CertificateAuthorityBackend) error {
	return nil
}

func (f *failingCaManager) UsedSecrets(string, *mesh_proto.CertificateAuthorityBackend) ([]string, error) {
	return nil, nil
}

func (f *failingCaManager) GetRootCert(context.Context, string, *mesh_proto.CertificateAuthorityBackend) ([]core_ca.Cert, error) {
	return nil, nil
}
