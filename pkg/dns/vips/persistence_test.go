package vips_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	mesh_core "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/manager"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"

	config_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

type countingConfigManager struct {
	create    int
	update    int
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
		err := rm.Create(context.Background(), mesh_core.NewMeshResource(), core_store.CreateByKey("mesh-1", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh_core.NewMeshResource(), core_store.CreateByKey("mesh-2", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())

		err = rm.Create(context.Background(), mesh_core.NewMeshResource(), core_store.CreateByKey("mesh-3", core_model.NoMesh))
		Expect(err).ToNot(HaveOccurred())
	})

	Context("Get", func() {
		BeforeEach(func() {
			meshedPersistence = vips.NewPersistence(rm, &countingConfigManager{
				configs: map[string]*system.ConfigResource{
					"kuma-mesh-1-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-1-dns-vips"},
						Spec: &config_proto.Config{Config: `{"backend":"240.0.0.1","frontend":"240.0.0.3","postgres":"240.0.0.0","redis":"240.0.0.2"}`},
					},
					"kuma-mesh-2-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
						Spec: &config_proto.Config{Config: `{"backend_2":"240.0.1.1","frontend_2":"240.0.1.3","postgres_2":"240.0.1.0","redis_2":"240.0.1.2"}`},
					},
					"kuma-mesh-3-dns-vips": {
						Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
						Spec: &config_proto.Config{Config: `{"backend_3":"240.0.2.1","frontend_3":"240.0.2.3","postgres_3":"240.0.2.0","redis_3":"240.0.2.2"}`},
					},
				},
			})
		})

		It("should merge vips from several configs", func() {
			actual, _, err := meshedPersistence.Get()
			Expect(err).ToNot(HaveOccurred())

			expected := vips.List{
				"backend":    "240.0.0.1",
				"backend_2":  "240.0.1.1",
				"backend_3":  "240.0.2.1",
				"frontend":   "240.0.0.3",
				"frontend_2": "240.0.1.3",
				"frontend_3": "240.0.2.3",
				"postgres":   "240.0.0.0",
				"postgres_2": "240.0.1.0",
				"postgres_3": "240.0.2.0",
				"redis":      "240.0.0.2",
				"redis_2":    "240.0.1.2",
				"redis_3":    "240.0.2.2",
			}
			Expect(actual).To(Equal(expected))
		})

		It("should return vips for mesh", func() {
			actual, err := meshedPersistence.GetByMesh("mesh-2")
			Expect(err).ToNot(HaveOccurred())
			expected := vips.List{
				"backend_2":  "240.0.1.1",
				"frontend_2": "240.0.1.3",
				"postgres_2": "240.0.1.0",
				"redis_2":    "240.0.1.2",
			}
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
			vipsMesh1 := vips.List{
				"backend": "240.0.0.1",
			}
			err := meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			vipsMesh2 := vips.List{
				"frontend": "240.0.0.2",
			}
			err = meshedPersistence.Set("mesh-2", vipsMesh2)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(2))
			Expect(countingCm.update).To(Equal(0))
		})

		It("should update existing config", func() {
			vipsMesh1 := vips.List{
				"backend": "240.0.0.1",
			}
			err := meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(1))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(0))

			vipsMesh1["frontend"] = "240.0.0.2"
			err = meshedPersistence.Set("mesh-1", vipsMesh1)
			Expect(err).ToNot(HaveOccurred())
			Expect(countingCm.get).To(Equal(2))
			Expect(countingCm.create).To(Equal(1))
			Expect(countingCm.update).To(Equal(1))
		})
	})
})
