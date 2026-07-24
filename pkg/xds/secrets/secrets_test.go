package secrets_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-retry"

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
	"github.com/kumahq/kuma/v3/pkg/core/xds/issuer"
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

		secrets, err = NewSecrets(caProvider, identityProvider, metrics, newTestLimiter(metrics, 5*time.Second))
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

		Context("when inbound tags are disabled", func() {
			newTaglessDataplane := func(labels map[string]string) *core_mesh.DataplaneResource {
				dp := newDataplane()
				dp.Spec.Networking.Inbound[0].Tags = map[string]string{}
				dp.Meta.(*model.ResourceMeta).Labels = labels
				return dp
			}

			It("should fall back to the kuma.io/workload label for identity", func() {
				// given a dataplane with no inbound tags but a workload label
				dataplane := newTaglessDataplane(map[string]string{"kuma.io/workload": "web"})

				// when
				identity, ca, err := secrets.GetForDataPlane(context.Background(), dataplane, newMesh("default"), nil)

				// then a cert is still generated, keyed off the workload label
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.PemCerts).ToNot(BeEmpty())
				Expect(ca).To(HaveLen(1))

				info := secrets.Info(mesh_proto.DataplaneProxyType, core_model.MetaToResourceKey(dataplane.Meta))
				Expect(info.Tags).To(Equal(mesh_proto.MultiValueTagSet{
					"kuma.io/service": map[string]bool{
						"web": true,
					},
				}))
			})

			It("should error rather than issue a cert with no SAN when the workload label is also missing", func() {
				// given a dataplane with neither inbound tags nor a workload label
				dataplane := newTaglessDataplane(nil)

				// when
				_, _, err := secrets.GetForDataPlane(context.Background(), dataplane, newMesh("default"), nil)

				// then
				Expect(err).To(HaveOccurred())
			})

			It("GetAllInOne should fall back to the kuma.io/workload label for identity", func() {
				// given a dataplane with no inbound tags but a workload label
				dataplane := newTaglessDataplane(map[string]string{"kuma.io/workload": "web"})

				// when
				identity, ca, err := secrets.GetAllInOne(context.Background(), newMesh("default"), dataplane, nil)

				// then a cert is still generated, keyed off the workload label
				Expect(err).ToNot(HaveOccurred())
				Expect(identity.PemCerts).ToNot(BeEmpty())
				Expect(ca.PemCerts).ToNot(BeEmpty())
			})

			It("GetAllInOne should error rather than issue a cert with no SAN when the workload label is also missing", func() {
				// given a dataplane with neither inbound tags nor a workload label
				dataplane := newTaglessDataplane(nil)

				// when
				_, _, err := secrets.GetAllInOne(context.Background(), newMesh("default"), dataplane, nil)

				// then
				Expect(err).To(HaveOccurred())
			})
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
		const backoffBase = 5 * time.Second

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
			// deterministic (no jitter) backoff so the test is stable
			failingSecrets, err = NewSecrets(caProvider, identityProvider, m, newTestLimiter(m, backoffBase))
			Expect(err).ToNot(HaveOccurred())
		})

		It("should not call the CA on every tick after a failure", func() {
			// when the first attempt fails
			_, _, err := failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)

			// then it starts a backoff and records a failure
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(1))
			Expect(test_metrics.FindMetric(failMetrics, "cert_generation_failure").GetCounter().GetValue()).To(Equal(1.0))
			Expect(test_metrics.FindMetric(failMetrics, "cert_generation_backoff").GetGauge().GetValue()).To(Equal(1.0))

			// when subsequent ticks happen within the backoff window
			for range 5 {
				_, _, err = failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)
				Expect(err).To(HaveOccurred())
			}

			// then the CA is not called again
			Expect(calls).To(Equal(1))

			// when the backoff window passes
			now = now.Add(backoffBase + time.Second)
			_, _, err = failingSecrets.GetForDataPlane(context.Background(), newDataplane(), failingMesh(), nil)

			// then the CA is retried
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(2))
		})
	})

	Context("does not serve stale certs while backing off", func() {
		dpCert := func() *mesh_proto.CertificateAuthorityBackend_DpCert {
			return &mesh_proto.CertificateAuthorityBackend_DpCert{
				Rotation: &mesh_proto.CertificateAuthorityBackend_DpCert_Rotation{Expiration: "1h"},
			}
		}
		// mesh with a working builtin backend (ca-1) and a failing one (ca-2)
		meshWith := func(enabled string) *core_mesh.MeshResource {
			return &core_mesh.MeshResource{
				Meta: &model.ResourceMeta{Name: "default"},
				Spec: &mesh_proto.Mesh{
					Mtls: &mesh_proto.Mesh_Mtls{
						EnabledBackend: enabled,
						Backends: []*mesh_proto.CertificateAuthorityBackend{
							{Name: "ca-1", Type: "builtin", DpCert: dpCert()},
							{Name: "ca-2", Type: "failing", DpCert: dpCert()},
						},
					},
				},
			}
		}

		It("returns an error instead of stale certs so the watchdog keeps retrying", func() {
			resStore := memory.NewStore()
			rm := manager.NewResourceManager(resStore)
			secretManager := secrets_manager.NewSecretManager(secrets_store.NewSecretStore(resStore), cipher.None(), nil, false)
			builtinCaManager := ca_builtin.NewBuiltinCaManager(secretManager)

			var calls int
			caManagers := core_ca.Managers{
				"builtin": builtinCaManager,
				"failing": &failingCaManager{calls: &calls},
			}

			mesh := meshWith("ca-1")
			Expect(rm.Create(context.Background(), mesh, store.CreateByKey(core_model.DefaultMesh, core_model.NoMesh))).To(Succeed())
			Expect(builtinCaManager.EnsureBackends(context.Background(), mesh, []*mesh_proto.CertificateAuthorityBackend{mesh.Spec.Mtls.Backends[0]})).To(Succeed())

			m, err := core_metrics.NewMetrics("local")
			Expect(err).ToNot(HaveOccurred())
			caProvider, err := NewCaProvider(caManagers, m)
			Expect(err).ToNot(HaveOccurred())
			identityProvider, err := NewIdentityProvider(caManagers, m)
			Expect(err).ToNot(HaveOccurred())
			s, err := NewSecrets(caProvider, identityProvider, m, newTestLimiter(m, 5*time.Second))
			Expect(err).ToNot(HaveOccurred())

			// given a DP with certs issued by the working backend
			identity, _, err := s.GetForDataPlane(context.Background(), newDataplane(), meshWith("ca-1"), nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(identity).ToNot(BeNil())

			// when the enabled backend switches to a failing one
			failing := meshWith("ca-2")
			_, _, err = s.GetForDataPlane(context.Background(), newDataplane(), failing, nil)

			// then generation fails and a backoff starts
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(1))

			// and while backing off it returns an error (not stale certs), so the
			// watchdog keeps its hash stale and keeps retrying; the CA is not hit again
			_, _, err = s.GetForDataPlane(context.Background(), newDataplane(), failing, nil)
			Expect(err).To(HaveOccurred())
			Expect(calls).To(Equal(1))
		})
	})
})

// newTestLimiter builds a limiter with a deterministic (no-jitter) per-proxy
// backoff and a circuit-breaker threshold above 1, so single-proxy tests
// exercise the backoff without ever tripping the backend circuit.
func newTestLimiter(m core_metrics.Metrics, backoff time.Duration) issuer.Limiter {
	l, err := issuer.NewLimiter(issuer.Config{
		NewBackoff: func() retry.Backoff { return retry.NewConstant(backoff) },
		MinProxies: 3,
		Window:     time.Minute,
		Cooldown:   30 * time.Second,
	}, m)
	Expect(err).ToNot(HaveOccurred())
	return l
}

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
