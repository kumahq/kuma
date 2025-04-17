package secrets_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core"
	core_ca "github.com/kumahq/kuma/pkg/core/ca"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/secrets/cipher"
	secrets_manager "github.com/kumahq/kuma/pkg/core/secrets/manager"
	secrets_store "github.com/kumahq/kuma/pkg/core/secrets/store"
	core_metrics "github.com/kumahq/kuma/pkg/metrics"
	ca_builtin "github.com/kumahq/kuma/pkg/plugins/ca/builtin"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_metrics "github.com/kumahq/kuma/pkg/test/metrics"
	"github.com/kumahq/kuma/pkg/test/resources/model"
	. "github.com/kumahq/kuma/pkg/xds/secrets"
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
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
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
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
		})

		Context("should regenerate certificate", func() {
			BeforeEach(func() {
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)
				Expect(err).ToNot(HaveOccurred())
			})

			FIt("when cross-mesh was added and previous removed", func() {
				// given
				defaultMesh := newMesh("default")

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), defaultMesh, []*core_mesh.MeshResource{newMesh("mesh-1")})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
				Expect(test_metrics.FindMetric(metrics, "cert_generation_other_meshes", "other_mesh", "mesh-1").GetCounter().GetValue()).To(Equal(1.0))

				// when
				_, _, err = secrets.GetForDataPlane(context.Background(), newDataplane(), defaultMesh, []*core_mesh.MeshResource{newMesh("mesh-2")})

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(1.0))
				Expect(test_metrics.FindMetric(metrics, "cert_generation_other_meshes", "other_mesh", "mesh-2").GetCounter().GetValue()).To(Equal(1.0))
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
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-2").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
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
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert is expiring", func() {
				// given
				now = now.Add(48*time.Minute + 1*time.Millisecond) // 4/5 of 60 minutes

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert was cleaned up", func() {
				// given
				secrets.Cleanup(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(newDataplane().Meta))

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), newDataplane(), newMesh("default"), nil)

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(2)))
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
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
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
			Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
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
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-2").GetSummary().GetSampleCount()).To(Equal(uint64(1)))
			})

			It("when cert is expiring", func() {
				// given
				now = now.Add(48*time.Minute + 1*time.Millisecond) // 4/5 of 60 minutes

				// when
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(2)))
			})

			It("when cert was cleaned up", func() {
				// given
				secrets.Cleanup(mesh_proto.EgressProxyType, core_model.MetaToResourceKey(newZoneEgress().Meta))

				// when
				_, _, err := secrets.GetForZoneEgress(context.Background(), newZoneEgress(), newMesh("default"))

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(test_metrics.FindMetric(metrics, "cert_generation").GetCounter().GetValue()).To(Equal(2.0))
				Expect(test_metrics.FindMetric(metrics, "ca_manager_get_cert", "backend_name", "ca-1").GetSummary().GetSampleCount()).To(Equal(uint64(2)))
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
})
