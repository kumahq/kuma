package persistence_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	config_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/dns/persistence"
	"github.com/kumahq/kuma/pkg/test/resources/model"
)

type testConfigManager struct {
	configs map[string]*system.ConfigResource
}

func (t *testConfigManager) Create(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.CreateOptionsFunc) error {
	panic("implement me")
}

func (t *testConfigManager) Update(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.UpdateOptionsFunc) error {
	panic("implement me")
}

func (t *testConfigManager) Delete(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.DeleteOptionsFunc) error {
	panic("implement me")
}

func (t *testConfigManager) DeleteAll(ctx context.Context, optionsFunc ...core_store.DeleteAllOptionsFunc) error {
	panic("implement me")
}

func (t *testConfigManager) Get(ctx context.Context, resource *system.ConfigResource, optionsFunc ...core_store.GetOptionsFunc) error {
	opts := core_store.NewGetOptions(optionsFunc...)
	resource.Spec.Config = t.configs[opts.Name].Spec.Config
	return nil
}

func (t *testConfigManager) List(ctx context.Context, list *system.ConfigResourceList, optionsFunc ...core_store.ListOptionsFunc) error {
	for _, i := range t.configs {
		err := list.AddItem(i)
		Expect(err).ToNot(HaveOccurred())
	}
	return nil
}

var _ = Describe("Meshed Persistence", func() {
	var p persistence.MeshedWriter
	//var cm manager.ConfigManager

	BeforeEach(func() {
		p = persistence.NewMeshedPersistence(&testConfigManager{
			configs: map[string]*system.ConfigResource{
				"kuma-mesh-1-dns-vips": {
					Meta: &model.ResourceMeta{Name: "kuma-mesh-1-dns-vips"},
					Spec: config_proto.Config{Config: `{"backend":"240.0.0.1","frontend":"240.0.0.3","postgres":"240.0.0.0","redis":"240.0.0.2"}`},
				},
				"kuma-mesh-2-dns-vips": {
					Meta: &model.ResourceMeta{Name: "kuma-mesh-2-dns-vips"},
					Spec: config_proto.Config{Config: `{"backend_2":"240.0.1.1","frontend_2":"240.0.1.3","postgres_2":"240.0.1.0","redis_2":"240.0.1.2"}`},
				},
				"kuma-mesh-3-dns-vips": {
					Meta: &model.ResourceMeta{Name: "kuma-mesh-3-dns-vips"},
					Spec: config_proto.Config{Config: `{"backend_3":"240.0.2.1","frontend_3":"240.0.2.3","postgres_3":"240.0.2.0","redis_3":"240.0.2.2"}`},
				},
			},
		})
	})

	It("should merge vips from several configs", func() {
		actual, err := p.Get()
		Expect(err).ToNot(HaveOccurred())

		expected := persistence.VIPList{
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
		actual, err := p.GetByMesh("mesh-2")
		Expect(err).ToNot(HaveOccurred())
		expected := persistence.VIPList{
			"backend_2":  "240.0.1.1",
			"frontend_2": "240.0.1.3",
			"postgres_2": "240.0.1.0",
			"redis_2":    "240.0.1.2",
		}
		Expect(actual).To(Equal(expected))
	})
})
