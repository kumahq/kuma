package vips_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

type countingConfigManager struct {
	create    int
	update    int
	updates   []*system.ConfigResource
	delete    int
	deleteAll int
	get       int
	list      int
	configs   map[string]*system.ConfigResource
}

func (t *countingConfigManager) Create(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.CreateOptionsFunc) error {
	t.create++
	opts := core_store.NewCreateOptions(optionsFunc...)
	mesh, ok := vips.MeshFromConfigKey(opts.Name)
	Expect(ok).To(BeTrue())
	Expect(mesh).To(Equal(opts.Owner.GetMeta().GetName()))
	t.configs[opts.Name] = resource
	return nil
}

func (t *countingConfigManager) Update(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.UpdateOptionsFunc) error {
	t.update++
	t.updates = append(t.updates, resource)
	return nil
}

func (t *countingConfigManager) Delete(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.DeleteOptionsFunc) error {
	t.delete++
	return nil
}

func (t *countingConfigManager) DeleteAll(ctx context.Context, optionsFunc ...core_store.DeleteAllOptionsFunc) error {
	t.deleteAll++
	return nil
}

func (t *countingConfigManager) Get(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.GetOptionsFunc) error {
	t.get++
	opts := core_store.NewGetOptions(optionsFunc...)
	if c, ok := t.configs[opts.Name]; ok {
		resource.Spec.Config = c.Spec.Config
		return nil
	}
	return core_store.ErrorResourceNotFound(system.ConfigType, opts.Name, core_model.NoMesh)
}

func (t *countingConfigManager) List(ctx context.Context, list *system.ConfigResourceList, optionsFunc ...core_store.ListOptionsFunc) error {
	t.list++
	for _, i := range t.configs {
		err := list.AddItem(i)
		Expect(err).ToNot(HaveOccurred())
	}
	return nil
}

var _ = Describe("Meshed Persistence", func() {
	var meshedPersistence *vips.Persistence
	var rm manager.ResourceManager

	BeforeEach(func() {
		rm = manager.NewResourceManager(memory.NewStore())
		err := rm.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-1", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-2", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey("mesh-3", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Get", func() {
		BeforeEach(func() {
			meshedPersistence = vips.NewPersistence(rm, &countingConfigManager{
				configs: map[string]*system.ConfigResource{
					"kuma-mesh-1-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-1-dns-vips"},
						Spec: &system_proto.Config{Config: `{"backend":"240.0.0.1","frontend":"240.0.0.3","postgres":"240.0.0.0","redis":"240.0.0.2"}`},
					},
					"kuma-mesh-2-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
						Spec: &system_proto.Config{Config: `{"backend_2":"240.0.1.1","frontend_2":"240.0.1.3","postgres_2":"240.0.1.0","redis_2":"240.0.1.2"}`},
					},
					"kuma-mesh-3-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
						Spec: &system_proto.Config{Config: `{"backend_3":"240.0.2.1","frontend_3":"240.0.2.3","postgres_3":"240.0.2.0","redis_3":"240.0.2.2"}`},
					},
				},
			})
		})

		It("should return vips for mesh", func() {
			actual, err := meshedPersistence.GetByMesh("mesh-2")
			Expect(err).ToNot(HaveOccurred())
			expected, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}, Origin: "legacy-service-upgrade"}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: "legacy-service-upgrade"}}},
				vips.NewServiceEntry("postgres_2"): {Address: "240.0.1.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres_2"}, Origin: "legacy-service-upgrade"}}},
				vips.NewServiceEntry("redis_2"):    {Address: "240.0.1.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis_2"}, Origin: "legacy-service-upgrade"}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})
	})

	Context("Set", func() {
		var countingCm *countingConfigManager

		BeforeEach(func() {
			countingCm = &countingConfigManager{
				configs: map[string]*system.ConfigResource{},
			}
			meshedPersistence = vips.NewPersistence(rm, countingCm)
		})

		It("should create a new config", func() {
			vipsMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{Port: 80, TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())

			err = meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			vipsMesh2, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set("mesh-2", vipsMesh2)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(2))
			Expect(countingCm.update).To(Equal(0))
		})

		It("should update existing config", func() {
			vipsMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			Expect(vipsMesh1.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{})).ToNot(HaveOccurred())
			err = meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(1))

			meshed, err := meshedPersistence.GetByMesh("mesh-1")
			Expect(err).ToNot(HaveOccurred())
			validateMeshes(meshed, "mesh-1")
		})
	})

	Context("Old and new configs at the same time", func() {
		var countingCm *countingConfigManager

		BeforeEach(func() {
			countingCm = &countingConfigManager{
				configs: map[string]*system.ConfigResource{
					"kuma-mesh-1-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-1-dns-vips"},
						Spec: &system_proto.Config{Config: `{"backend":"240.0.0.1","frontend":"240.0.0.3","postgres":"240.0.0.0","redis":"240.0.0.2"}`},
					},
					"kuma-mesh-2-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
						Spec: &system_proto.Config{Config: `{"0:backend_2":"240.0.1.1","0:frontend_2":"240.0.1.3","1:host.com":"240.0.1.4"}`},
					},
					"kuma-mesh-3-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
						Spec: &system_proto.Config{Config: `{"0:backend_3":{"address":"240.0.2.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_3"}}]},"0:frontend_3":{"address":"240.0.2.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_3"}}]},"1:host.com":{"address":"240.0.1.4","outbounds":[{"TagSet":{"kuma.io/service":"external-host"}}]}}`},
					},
				},
			}
			_ = rm.Create(context.Background(), &core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "host.com:90",
					},
					Tags: map[string]string{mesh_proto.ServiceTag: "external-host"},
				},
			}, core_store.CreateByKey("my-host", "mesh-2"))
			_ = rm.Create(context.Background(), &core_mesh.ExternalServiceResource{
				Spec: &mesh_proto.ExternalService{
					Networking: &mesh_proto.ExternalService_Networking{
						Address: "host.com:100",
					},
					Tags: map[string]string{mesh_proto.ServiceTag: "external-host2"},
				},
			}, core_store.CreateByKey("my-host2", "mesh-2"))
			meshedPersistence = vips.NewPersistence(rm, countingCm)
		})

		It("should return meshed vips", func() {
			// when
			meshed, err := meshedPersistence.Get()
			Expect(err).ToNot(HaveOccurred())
			for mesh, v := range meshed {
				validateMeshes(v, mesh)
			}

			// and then
			expectedMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
				vips.NewServiceEntry("postgres"): {Address: "240.0.0.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres"}}}},
				vips.NewServiceEntry("redis"):    {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(meshed["mesh-1"].HostnameEntries()).To(Equal(expectedMesh1.HostnameEntries()))
			// and then
			expectedMesh2, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}}}},
				vips.NewHostEntry("host.com"):      {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(meshed["mesh-2"].HostnameEntries()).To(Equal(expectedMesh2.HostnameEntries()))
			// and then
			expectedMesh3, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_3"):  {Address: "240.0.2.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_3"}}}},
				vips.NewServiceEntry("frontend_3"): {Address: "240.0.2.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_3"}}}},
				vips.NewHostEntry("host.com"):      {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(meshed["mesh-3"]).To(WithTransform(func(vo *vips.VirtualOutboundMeshView) []vips.HostnameEntry {
				return vo.HostnameEntries()
			}, Equal(expectedMesh3.HostnameEntries())))
			for _, k := range meshed["mesh-3"].HostnameEntries() {
				Expect(meshed["mesh-3"].Get(k)).To(Equal(expectedMesh3.Get(k)))
			}
		})

		It("should update old version with new one", func() {
			newVIPs, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
				vips.NewHostEntry("kuma.io"):     {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set("mesh-1", newVIPs)
			Expect(err).ToNot(HaveOccurred())

			Expect(countingCm.updates).To(HaveLen(1))
			Expect(countingCm.updates[0].Spec.Config).To(Equal(`{"0:backend":{"address":"240.0.0.1","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"backend"},"Origin":""}]},"0:frontend":{"address":"240.0.0.3","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"frontend"},"Origin":""}]},"1:kuma.io":{"address":"240.0.1.4","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"external-host"},"Origin":""}]}}`))
		})

		It("should handle multiple external service with same host", func() {
			out, err := meshedPersistence.GetByMesh("mesh-2")
			Expect(err).ToNot(HaveOccurred())
			entry := out.Get(vips.NewHostEntry("host.com"))

			Expect(entry.Outbounds).To(Equal([]vips.OutboundEntry{
				{Port: 90, TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}, Origin: "legacy-host-upgrade"},
				{Port: 100, TagSet: map[string]string{mesh_proto.ServiceTag: "external-host2"}, Origin: "legacy-host-upgrade"},
			}))
		})
	})
})

func validateMeshes(v *vips.VirtualOutboundMeshView, meshName string) {
	for _, hostname := range v.HostnameEntries() {
		vo := v.Get(hostname)
		Expect(vo).ToNot(BeNil(), fmt.Sprintf("mesh: %s, hostname: %s", meshName, hostname))
		Expect(vo.Outbounds).ToNot(BeEmpty(), fmt.Sprintf("mesh: %s, hostname: %s", meshName, hostname))
	}
}
