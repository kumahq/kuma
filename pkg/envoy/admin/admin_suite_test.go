package admin_test

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	core_mesh "github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/defaults/mesh"
	"github.com/kumahq/kuma/pkg/envoy/admin"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/runtime"
)

func TestEnvoyAdmin(t *testing.T) {
	test.RunSpecs(t, "EnvoyAdmin Suite")
}

const (
	testMesh    = "test-mesh"
	anotherMesh = "another-mesh"
)

var eac admin.EnvoyAdminClient

var _ = BeforeSuite(func() {

	// setup the runtime
	cfg := kuma_cp.DefaultConfig()
	builder, err := runtime.BuilderFor(context.Background(), cfg)
	Expect(err).ToNot(HaveOccurred())
	runtime, err := builder.Build()
	Expect(err).ToNot(HaveOccurred())
	resManager := runtime.ResourceManager()
	Expect(resManager).ToNot(BeNil())

	// create mesh defaults
	err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(testMesh, model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	err = mesh.EnsureDefaultMeshResources(runtime.ResourceManager(), testMesh)
	Expect(err).ToNot(HaveOccurred())

	err = resManager.Create(context.Background(), core_mesh.NewMeshResource(), core_store.CreateByKey(anotherMesh, model.NoMesh))
	Expect(err).ToNot(HaveOccurred())

	err = mesh.EnsureDefaultMeshResources(runtime.ResourceManager(), anotherMesh)
	Expect(err).ToNot(HaveOccurred())

	// setup the Envoy Admin Client
	eac = admin.NewEnvoyAdminClient(resManager, runtime.Config())
	Expect(eac).ToNot(BeNil())
})
