package logs_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/logs"
	core_mesh "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
	core_manager "github.com/Kong/kuma/pkg/core/resources/manager"
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	"github.com/Kong/kuma/pkg/util/proto"
)

var _ = Describe("Matcher", func() {

	var manager core_manager.ResourceManager
	var matcher logs.TrafficLogsMatcher
	var dpRes core_mesh.DataplaneResource

	var backendFile1 *mesh_proto.LoggingBackend
	var backendFile2 *mesh_proto.LoggingBackend
	var backendFile3 *mesh_proto.LoggingBackend

	BeforeEach(func() {
		manager = core_manager.NewResourceManager(memory.NewStore())
		matcher = logs.TrafficLogsMatcher{manager}

		// given mesh with 3 backends and file1 backend as default
		backendFile1 = &mesh_proto.LoggingBackend{
			Name: "file1",
			Type: mesh_proto.LoggingFileType,
			Conf: proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
				Path: "/tmp/access.logs",
			}),
		}
		backendFile2 = &mesh_proto.LoggingBackend{
			Name: "file2",
			Type: mesh_proto.LoggingFileType,
			Conf: proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
				Path: "/tmp/access.logs",
			}),
		}
		backendFile3 = &mesh_proto.LoggingBackend{
			Name: "file3",
			Type: mesh_proto.LoggingFileType,
			Conf: proto.MustToStruct(&mesh_proto.FileLoggingBackendConfig{
				Path: "/tmp/access.logs",
			}),
		}
		meshRes := core_mesh.MeshResource{
			Spec: mesh_proto.Mesh{
				Logging: &mesh_proto.Logging{
					Backends:       []*mesh_proto.LoggingBackend{backendFile1, backendFile2, backendFile3},
					DefaultBackend: "file1",
				},
			},
		}
		err := manager.Create(context.Background(), &meshRes, store.CreateByKey("sample", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// and
		dpRes = core_mesh.DataplaneResource{
			Spec: mesh_proto.Dataplane{
				Networking: &mesh_proto.Dataplane_Networking{
					Address: "127.0.0.1",
					Inbound: []*mesh_proto.Dataplane_Networking_Inbound{
						{
							Port:        8080,
							ServicePort: 8081,
							Tags: map[string]string{
								"service": "kong",
							},
						},
						{
							Port:        8090,
							ServicePort: 8091,
							Tags: map[string]string{
								"service": "kong-admin",
							},
						},
					},
					Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
						{
							Port:    9091,
							Service: "backend",
						},
						{
							Port:    9092,
							Service: "web",
						},
					},
				},
			},
		}
		err = manager.Create(context.Background(), &dpRes, store.CreateByKey("dp-1", "sample"))
		Expect(err).ToNot(HaveOccurred())
	})

	It("should match rules", func() {
		// given
		logRes1 := core_mesh.TrafficLogResource{
			Spec: mesh_proto.TrafficLog{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "kong",
						},
					},
				},
				Destinations: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
				Conf: &mesh_proto.TrafficLog_Conf{
					Backend: "file2",
				},
			},
		}
		err := manager.Create(context.Background(), &logRes1, store.CreateByKey("lr-1", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// and
		logRes3 := core_mesh.TrafficLogResource{
			Spec: mesh_proto.TrafficLog{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "*",
						},
					},
				},
				Destinations: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "*",
						},
					},
				},
			},
		}
		err = manager.Create(context.Background(), &logRes3, store.CreateByKey("lr-3", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// when
		log, err := matcher.Match(context.Background(), &dpRes)

		// then
		Expect(err).ToNot(HaveOccurred())
		// should match because kong->backend rule
		Expect(log["backend"]).To(Equal(backendFile2))
		// should match because *->* rule and default backend file1
		Expect(log["web"]).To(Equal(backendFile1))
		// should match implicit pass through because service *->* rule and default backend file1
		Expect(log[core_mesh.PassThroughService]).To(Equal(backendFile1))
	})

	It("should not match services", func() {
		// given
		logRes := core_mesh.TrafficLogResource{
			Spec: mesh_proto.TrafficLog{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "web",
						},
					},
				},
				Destinations: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "backend",
						},
					},
				},
				Conf: &mesh_proto.TrafficLog_Conf{
					Backend: "file2",
				},
			},
		}
		err := manager.Create(context.Background(), &logRes, store.CreateByKey("lr-1", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// when
		log, err := matcher.Match(context.Background(), &dpRes)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(log).To(HaveLen(0))
	})

	It("should skip unknown backends", func() {
		// given
		logRes := core_mesh.TrafficLogResource{
			Spec: mesh_proto.TrafficLog{
				Sources: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "*",
						},
					},
				},
				Destinations: []*mesh_proto.Selector{
					{
						Match: map[string]string{
							"service": "*",
						},
					},
				},
				Conf: &mesh_proto.TrafficLog_Conf{
					Backend: "unknown-backend",
				},
			},
		}
		err := manager.Create(context.Background(), &logRes, store.CreateByKey("lr-1", "sample"))
		Expect(err).ToNot(HaveOccurred())

		// when
		log, err := matcher.Match(context.Background(), &dpRes)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(log).To(HaveLen(0))
	})
})
