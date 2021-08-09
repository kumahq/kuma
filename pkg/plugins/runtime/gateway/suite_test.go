package gateway_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	"github.com/kumahq/kuma/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/core/resources/model/rest"
	"github.com/kumahq/kuma/pkg/core/resources/registry"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/core/runtime"
	"github.com/kumahq/kuma/pkg/test"
	test_runtime "github.com/kumahq/kuma/pkg/test/runtime"
)

func TestGateway(t *testing.T) {
	test.RunSpecs(t, "Gateway Suite")
}

// FetchNamedFixture retrieves the named resource from the runtime
// resource manager.
func FetchNamedFixture(
	rt runtime.Runtime,
	resourceType core_model.ResourceType,
	key core_model.ResourceKey,
) (core_model.Resource, error) {
	r, err := registry.Global().NewObject(resourceType)
	if err != nil {
		return nil, err
	}

	if err := rt.ReadOnlyResourceManager().Get(context.TODO(), r, store.GetBy(key)); err != nil {
		return nil, err
	}

	return r, nil
}

// StoreNamedFixture reads the given YAML file name from the testdata
// directory, then stores it in the runtime resource manager.
func StoreNamedFixture(rt runtime.Runtime, name string) error {
	bytes, err := ioutil.ReadFile(path.Join("testdata", name))
	if err != nil {
		return err
	}

	r, err := rest.UnmarshallToCore(bytes)
	if err != nil {
		return err
	}

	var opts []store.CreateOptionsFunc

	switch r.Descriptor().Scope {
	case core_model.ScopeGlobal:
		opts = append(opts, store.CreateByKey(r.GetMeta().GetName(), ""))
	case core_model.ScopeMesh:
		opts = append(opts, store.CreateByKey(r.GetMeta().GetName(), r.GetMeta().GetMesh()))
	}

	return rt.ResourceManager().Create(context.TODO(), r, opts...)
}

// BuildRuntime returns a fabricated test Runtime instance with which
// the gateway plugin is registered.
func BuildRuntime() (runtime.Runtime, error) {
	builder, err := test_runtime.BuilderFor(kuma_cp.DefaultConfig())
	if err != nil {
		return nil, err
	}

	rt, err := builder.Build()
	if err != nil {
		return nil, err
	}

	if err := plugins.Plugins().RuntimePlugins()["gateway"].Customize(rt); err != nil {
		return nil, err
	}

	return rt, nil
}

var _ = BeforeSuite(func() {
	// Ensure that the plugin is registered so that tests at least
	// have a chance of working.
	_, registered := plugins.Plugins().RuntimePlugins()["gateway"]
	Expect(registered).To(BeTrue(), "gateway plugin is registered")
})
