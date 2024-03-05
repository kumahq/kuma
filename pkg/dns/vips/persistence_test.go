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
						Spec: &system_proto.Config{Config: `{"0:backend":{"address":"240.0.0.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_1"}}]},"0:frontend":{"address":"240.0.0.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_2"}}]},"0:redis":{"address":"240.0.0.2","outbounds":[{"TagSet":{"kuma.io/service":"redis"}}]},"0:postgres":{"address":"240.0.0.0","outbounds":[{"TagSet":{"kuma.io/service":"postgres"}}]}}`},
					},
					"kuma-mesh-2-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
						Spec: &system_proto.Config{Config: `{"0:backend_2":{"address":"240.0.1.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_2"}}]},"0:frontend_2":{"address":"240.0.1.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_2"}}]},"0:postgres_2":{"address":"240.0.1.0","outbounds":[{"TagSet":{"kuma.io/service":"postgres_2"}}]},"0:redis_2":{"address":"240.0.1.2","outbounds":[{"TagSet":{"kuma.io/service":"redis_2"}}]}}`},
					},
					"kuma-mesh-3-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
						Spec: &system_proto.Config{Config: `{"0:backend_3":{"address":"240.0.2.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_3"}}]},"0:frontend_3":{"address":"240.0.2.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_3"}}]},"1:host.com":{"address":"240.0.1.4","outbounds":[{"TagSet":{"kuma.io/service":"external-host"}}]}}`},
					},
					"kuma-mesh-4-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-4-dns-vips"},
						Spec: &system_proto.Config{Config: `{"v":{"backend_2":{"oe":[{"ap":[{"a":"240.0.1.1","o":"svc"}]}]},"frontend_2":{"oe":[{"ap":[{"a":"240.0.1.3","o":"svc"}]}]},"postgres_2":{"oe":[{"ap":[{"a":"240.0.1.0","o":"svc"}]}]}}}`},
					},
				},
			}, false)
		})

		It("should return vips for mesh", func() {
			actual, err := meshedPersistence.GetByMesh(context.Background(), "mesh-2")
			Expect(err).ToNot(HaveOccurred())
			expected, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}, Origin: ""}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: ""}}},
				vips.NewServiceEntry("postgres_2"): {Address: "240.0.1.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres_2"}, Origin: ""}}},
				vips.NewServiceEntry("redis_2"):    {Address: "240.0.1.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis_2"}, Origin: ""}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})

		It("should return vips for mesh in new format", func() {
			actual, err := meshedPersistence.GetByMesh(context.Background(), "mesh-4")
			Expect(err).ToNot(HaveOccurred())
			expected, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}, Origin: "service"}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: "service"}}},
				vips.NewServiceEntry("postgres_2"): {Address: "240.0.1.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres_2"}, Origin: "service"}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(Equal(expected))
		})

		It("should return vips for all meshes", func() {
			actual, err := meshedPersistence.Get(context.Background(), []string{"mesh-1", "mesh-2", "mesh-3", "mesh-4"})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual).To(HaveLen(4))

			// mesh-1
			expected, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_1"}, Origin: ""}}},
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: ""}}},
				vips.NewServiceEntry("redis"):    {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis"}, Origin: ""}}},
				vips.NewServiceEntry("postgres"): {Address: "240.0.0.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres"}, Origin: ""}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual["mesh-1"]).To(Equal(expected))

			// mesh-2
			expected, err = vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}, Origin: ""}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: ""}}},
				vips.NewServiceEntry("postgres_2"): {Address: "240.0.1.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres_2"}, Origin: ""}}},
				vips.NewServiceEntry("redis_2"):    {Address: "240.0.1.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis_2"}, Origin: ""}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual["mesh-2"]).To(Equal(expected))

			// mesh-3
			expected, err = vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_3"):  {Address: "240.0.2.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_3"}, Origin: ""}}},
				vips.NewServiceEntry("frontend_3"): {Address: "240.0.2.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_3"}, Origin: ""}}},
				vips.NewHostEntry("host.com"):      {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}, Origin: ""}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual["mesh-3"]).To(Equal(expected))

			// mesh-4
			expected, err = vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}, Origin: "service"}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}, Origin: "service"}}},
				vips.NewServiceEntry("postgres_2"): {Address: "240.0.1.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres_2"}, Origin: "service"}}},
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(actual["mesh-4"]).To(Equal(expected))
		})
	})

	Context("Set", func() {
		var countingCm *countingConfigManager

		BeforeEach(func() {
			countingCm = &countingConfigManager{
				configs: map[string]*system.ConfigResource{},
			}
			meshedPersistence = vips.NewPersistence(rm, countingCm, false)
		})

		It("should create a new config", func() {
			vipsMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{Port: 80, TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())

			err = meshedPersistence.Set(context.Background(), "mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			vipsMesh2, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set(context.Background(), "mesh-2", vipsMesh2)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(2))
			Expect(countingCm.update).To(Equal(0))
		})

		It("should update existing config", func() {
			ctx := context.Background()
			vipsMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set(ctx, "mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			Expect(vipsMesh1.Add(vips.NewServiceEntry("frontend"), vips.OutboundEntry{})).ToNot(HaveOccurred())
			err = meshedPersistence.Set(ctx, "mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(1))

			meshed, err := meshedPersistence.GetByMesh(ctx, "mesh-1")
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
						Spec: &system_proto.Config{Config: `{
								"0:backend":{"address":"240.0.0.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_1"}}]},
								"0:frontend":{"address":"240.0.0.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_2"}}]},
								"0:redis":{"address":"240.0.0.2","outbounds":[{"TagSet":{"kuma.io/service":"redis"}}]},
								"0:postgres":{"address":"240.0.0.0","outbounds":[{"TagSet":{"kuma.io/service":"postgres"}}]}}`},
					},
					"kuma-mesh-2-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
						Spec: &system_proto.Config{Config: `{
									"0:backend_2":{"address":"240.0.1.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_2"}}]},
									"0:frontend_2":{"address":"240.0.1.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_2"}}]},
									"1:host.com":{"address":"240.0.1.4","outbounds":[
										{"Port": 90, "TagSet":{"kuma.io/service":"external-host"}},
										{"Port": 100, "TagSet":{"kuma.io/service":"external-host2"}}
									]}
						}`},
					},
					"kuma-mesh-3-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
						Spec: &system_proto.Config{Config: `{"0:backend_3":{"address":"240.0.2.1","outbounds":[{"TagSet":{"kuma.io/service":"backend_3"}}]},"0:frontend_3":{"address":"240.0.2.3","outbounds":[{"TagSet":{"kuma.io/service":"frontend_3"}}]},"1:host.com":{"address":"240.0.1.4","outbounds":[{"TagSet":{"kuma.io/service":"external-host"}}]}}`},
					},
				},
			}
			meshedPersistence = vips.NewPersistence(rm, countingCm, false)
		})

		It("should return meshed vips", func() {
			expectedMesh1, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
				vips.NewServiceEntry("postgres"): {Address: "240.0.0.0", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "postgres"}}}},
				vips.NewServiceEntry("redis"):    {Address: "240.0.0.2", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "redis"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			view, err := meshedPersistence.GetByMesh(context.Background(), "mesh-1")
			Expect(err).ToNot(HaveOccurred())
			validateMeshes(view, "mesh-1")
			Expect(view.HostnameEntries()).To(Equal(expectedMesh1.HostnameEntries()))

			// and then
			expectedMesh2, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_2"):  {Address: "240.0.1.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_2"}}}},
				vips.NewServiceEntry("frontend_2"): {Address: "240.0.1.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_2"}}}},
				vips.NewHostEntry("host.com"):      {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			view, err = meshedPersistence.GetByMesh(context.Background(), "mesh-2")
			Expect(err).ToNot(HaveOccurred())
			validateMeshes(view, "mesh-2")
			Expect(view.HostnameEntries()).To(Equal(expectedMesh2.HostnameEntries()))

			// and then
			expectedMesh3, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend_3"):  {Address: "240.0.2.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend_3"}}}},
				vips.NewServiceEntry("frontend_3"): {Address: "240.0.2.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend_3"}}}},
				vips.NewHostEntry("host.com"):      {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			view, err = meshedPersistence.GetByMesh(context.Background(), "mesh-3")
			Expect(err).ToNot(HaveOccurred())
			validateMeshes(view, "mesh-3")
			Expect(view).To(WithTransform(func(vo *vips.VirtualOutboundMeshView) []vips.HostnameEntry {
				return vo.HostnameEntries()
			}, Equal(expectedMesh3.HostnameEntries())))
			for _, k := range view.HostnameEntries() {
				Expect(view.Get(k)).To(Equal(expectedMesh3.Get(k)))
			}
		})

		It("should update", func() {
			newVIPs, err := vips.NewVirtualOutboundView(map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("backend"):  {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "backend"}}}},
				vips.NewServiceEntry("frontend"): {Address: "240.0.0.3", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "frontend"}}}},
				vips.NewHostEntry("kuma.io"):     {Address: "240.0.1.4", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}}}},
			})
			Expect(err).ToNot(HaveOccurred())
			err = meshedPersistence.Set(context.Background(), "mesh-1", newVIPs)
			Expect(err).ToNot(HaveOccurred())

			Expect(countingCm.updates).To(HaveLen(1))
			Expect(countingCm.updates[0].Spec.Config).To(Equal(`{"0:backend":{"address":"240.0.0.1","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"backend"},"Origin":""}]},"0:frontend":{"address":"240.0.0.3","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"frontend"},"Origin":""}]},"1:kuma.io":{"address":"240.0.1.4","outbounds":[{"Port":0,"TagSet":{"kuma.io/service":"external-host"},"Origin":""}]}}`))
		})

		It("should handle multiple external service with same host", func() {
			out, err := meshedPersistence.GetByMesh(context.Background(), "mesh-2")
			Expect(err).ToNot(HaveOccurred())
			entry := out.Get(vips.NewHostEntry("host.com"))

			Expect(entry.Outbounds).To(Equal([]vips.OutboundEntry{
				{Port: 90, TagSet: map[string]string{mesh_proto.ServiceTag: "external-host"}, Origin: ""},
				{Port: 100, TagSet: map[string]string{mesh_proto.ServiceTag: "external-host2"}, Origin: ""},
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
