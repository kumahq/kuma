package externalservice

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	ratelimit_managers "github.com/kumahq/kuma/pkg/core/managers/apis/ratelimit"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/validators"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("ExternalService Manager", func() {

	var resStore store.ResourceStore
	var rateLimitManager manager.ResourceManager
	var externalServiceManager manager.ResourceManager

	BeforeEach(func() {
		resStore = memory.NewStore()
		rateLimitValidator := ratelimit_managers.RateLimitValidator{
			Store: resStore,
		}
		rateLimitManager = ratelimit_managers.NewRateLimitManager(resStore, rateLimitValidator)

		externalServiceValidator := ExternalServiceValidator{
			Store: resStore,
		}
		externalServiceManager = NewExternalServiceManager(resStore, externalServiceValidator)
	})

	Describe("Create()", func() {
		It("Should allow outbound with only service tag", func() {
			// given
			meshName := "mesh-1"
			rlKey := model.ResourceKey{
				Mesh: meshName,
				Name: "ratelimit1",
			}

			esKey := model.ResourceKey{
				Mesh: meshName,
				Name: "service2",
			}

			// when
			externalService := core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "example.com:80",
					},
					Tags: map[string]string{
						"kuma.io/service": "service2",
					},
				},
			}
			ratelimit := core_mesh.RateLimitResource{
				Spec: &mesh_proto.RateLimit{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service1",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service2",
							},
						},
					},
					Conf: &mesh_proto.RateLimit_Conf{
						Http: &mesh_proto.RateLimit_Conf_Http{
							Requests: 100,
							Interval: util_proto.Duration(time.Second * 10),
						},
					},
				},
			}
			err := rateLimitManager.Create(context.Background(), &ratelimit, store.CreateBy(rlKey))
			Expect(err).ToNot(HaveOccurred())

			err = externalServiceManager.Create(context.Background(), &externalService, store.CreateBy(esKey))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should allow outbound with only service tag = *", func() {
			// given
			meshName := "mesh-1"
			rlKey := model.ResourceKey{
				Mesh: meshName,
				Name: "ratelimit1",
			}

			esKey := model.ResourceKey{
				Mesh: meshName,
				Name: "service2-name",
			}

			// when
			externalService := core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "example.com:80",
					},
					Tags: map[string]string{
						"kuma.io/service": "service2",
					},
				},
			}
			ratelimit := core_mesh.RateLimitResource{
				Spec: &mesh_proto.RateLimit{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service1",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "*",
							},
						},
					},
					Conf: &mesh_proto.RateLimit_Conf{
						Http: &mesh_proto.RateLimit_Conf_Http{
							Requests: 100,
							Interval: util_proto.Duration(time.Second * 10),
						},
					},
				},
			}
			err := rateLimitManager.Create(context.Background(), &ratelimit, store.CreateBy(rlKey))
			Expect(err).ToNot(HaveOccurred())

			err = externalServiceManager.Create(context.Background(), &externalService, store.CreateBy(esKey))

			// then
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should disallow outbound with service tag + other tag", func() {
			// given
			meshName := "mesh-1"
			rlKey := model.ResourceKey{
				Mesh: meshName,
				Name: "ratelimit1",
			}

			esKey := model.ResourceKey{
				Mesh: meshName,
				Name: "service2-name",
			}

			// when
			externalService := core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "example.com:80",
					},
					Tags: map[string]string{
						"kuma.io/service": "service2",
					},
				},
			}
			ratelimit := core_mesh.RateLimitResource{
				Spec: &mesh_proto.RateLimit{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service1",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service2",
								"version":         "v1",
							},
						},
					},
					Conf: &mesh_proto.RateLimit_Conf{
						Http: &mesh_proto.RateLimit_Conf_Http{
							Requests: 100,
							Interval: util_proto.Duration(time.Second * 10),
						},
					},
				},
			}
			err := rateLimitManager.Create(context.Background(), &ratelimit, store.CreateBy(rlKey))
			Expect(err).ToNot(HaveOccurred())

			err = externalServiceManager.Create(context.Background(), &externalService, store.CreateBy(esKey))

			// then
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "externalservice",
						Message: "ExternalService would match RateLimit 'ratelimit1' which includes incompatible destination matches",
					},
				},
			}))
		})
	})

	Describe("Update()", func() {
		It("Should allow outbound with only service tag", func() {
			// given
			meshName := "mesh-1"
			rlKey := model.ResourceKey{
				Mesh: meshName,
				Name: "ratelimit1",
			}

			esKey := model.ResourceKey{
				Mesh: meshName,
				Name: "service2",
			}

			// when
			externalService := core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "example.com:80",
					},
					Tags: map[string]string{
						"kuma.io/service": "service3",
					},
				},
			}
			ratelimit := core_mesh.RateLimitResource{
				Spec: &mesh_proto.RateLimit{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service1",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service2",
							},
						},
					},
					Conf: &mesh_proto.RateLimit_Conf{
						Http: &mesh_proto.RateLimit_Conf_Http{
							Requests: 100,
							Interval: util_proto.Duration(time.Second * 10),
						},
					},
				},
			}
			err := rateLimitManager.Create(context.Background(), &ratelimit, store.CreateBy(rlKey))
			Expect(err).ToNot(HaveOccurred())

			err = externalServiceManager.Create(context.Background(), &externalService, store.CreateBy(esKey))
			Expect(err).ToNot(HaveOccurred())

			// then
			externalService.Spec.Tags["kuma.io/service"] = "service2"
			err = externalServiceManager.Update(context.Background(), &externalService)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Should disallow outbound with service tag + other tag", func() {
			// given
			meshName := "mesh-1"
			rlKey := model.ResourceKey{
				Mesh: meshName,
				Name: "ratelimit1",
			}

			esKey := model.ResourceKey{
				Mesh: meshName,
				Name: "service2-name",
			}

			// when
			externalService := core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "example.com:80",
					},
					Tags: map[string]string{
						"kuma.io/service": "service2",
					},
				},
			}
			ratelimit := core_mesh.RateLimitResource{
				Spec: &mesh_proto.RateLimit{
					Sources: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service1",
							},
						},
					},
					Destinations: []*mesh_proto.Selector{
						{
							Match: map[string]string{
								"kuma.io/service": "service3",
								"version":         "v1",
							},
						},
					},
					Conf: &mesh_proto.RateLimit_Conf{
						Http: &mesh_proto.RateLimit_Conf_Http{
							Requests: 100,
							Interval: util_proto.Duration(time.Second * 10),
						},
					},
				},
			}
			err := externalServiceManager.Create(context.Background(), &externalService, store.CreateBy(esKey))
			Expect(err).ToNot(HaveOccurred())

			err = rateLimitManager.Create(context.Background(), &ratelimit, store.CreateBy(rlKey))
			Expect(err).ToNot(HaveOccurred())

			// then
			externalService.Spec.Tags["kuma.io/service"] = "service3"
			err = externalServiceManager.Update(context.Background(), &externalService)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(&validators.ValidationError{
				Violations: []validators.Violation{
					{
						Field:   "externalservice",
						Message: "ExternalService would match RateLimit 'ratelimit1' which includes incompatible destination matches",
					},
				},
			}))
		})

	})

})
