package defaults_test

import (
	"context"
	"slices"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/pkg/config/core"
	config_store "github.com/kumahq/kuma/pkg/config/core/resources/store"
	"github.com/kumahq/kuma/pkg/config/plugins/resources/k8s"
	hostnamegenerator_api "github.com/kumahq/kuma/pkg/core/resources/apis/hostnamegenerator/api/v1alpha1"
	core_manager "github.com/kumahq/kuma/pkg/core/resources/manager"
	"github.com/kumahq/kuma/pkg/defaults"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
)

var _ = Describe("Ensure Hostname Generators", func() {
	type testCase struct {
		cpConfig         kuma_cp.Config
		expectedGenNames []string
	}

	DescribeTable("should create default generators",
		func(given testCase) {
			// given
			resManager := core_manager.NewResourceManager(memory.NewStore())

			// when
			err := defaults.EnsureHostnameGeneratorExists(
				context.Background(),
				resManager,
				logr.Discard(),
				given.cpConfig,
			)

			// then
			Expect(err).ToNot(HaveOccurred())

			generators := hostnamegenerator_api.HostnameGeneratorResourceList{}
			Expect(resManager.List(context.Background(), &generators)).To(Succeed())
			var createdNames []string
			for _, gen := range generators.Items {
				createdNames = append(createdNames, gen.GetMeta().GetName())
			}
			slices.Sort(createdNames)
			Expect(createdNames).To(Equal(given.expectedGenNames))
		},
		Entry("skip defaults", testCase{
			cpConfig: kuma_cp.Config{
				Defaults: &kuma_cp.Defaults{
					SkipHostnameGenerators: true,
				},
			},
			expectedGenNames: nil,
		}),
		Entry("global universal", testCase{
			cpConfig: kuma_cp.Config{
				Defaults:    &kuma_cp.Defaults{},
				Mode:        config_core.Global,
				Environment: config_core.UniversalEnvironment,
			},
			expectedGenNames: []string{
				"synced-headless-kube-mesh-service",
				"synced-kube-mesh-service",
				"synced-mesh-external-service",
				"synced-mesh-multi-zone-service",
				"synced-universal-mesh-service",
			},
		}),
		Entry("global kubernetes", testCase{
			cpConfig: kuma_cp.Config{
				Defaults:    &kuma_cp.Defaults{},
				Mode:        config_core.Global,
				Environment: config_core.KubernetesEnvironment,
				Store: &config_store.StoreConfig{
					Kubernetes: &k8s.KubernetesStoreConfig{
						SystemNamespace: "kuma-system",
					},
				},
			},
			expectedGenNames: []string{
				"synced-headless-kube-mesh-service.kuma-system",
				"synced-kube-mesh-service.kuma-system",
				"synced-mesh-external-service.kuma-system",
				"synced-mesh-multi-zone-service.kuma-system",
				"synced-universal-mesh-service.kuma-system",
			},
		}),
		Entry("zone kubernetes", testCase{
			cpConfig: kuma_cp.Config{
				Defaults:    &kuma_cp.Defaults{},
				Mode:        config_core.Zone,
				Environment: config_core.KubernetesEnvironment,
				Store: &config_store.StoreConfig{
					Kubernetes: &k8s.KubernetesStoreConfig{
						SystemNamespace: "kuma-system",
					},
				},
			},
			expectedGenNames: []string{
				"local-mesh-external-service.kuma-system",
			},
		}),
		Entry("zone universal", testCase{
			cpConfig: kuma_cp.Config{
				Defaults:    &kuma_cp.Defaults{},
				Mode:        config_core.Zone,
				Environment: config_core.UniversalEnvironment,
			},
			expectedGenNames: []string{
				"local-mesh-external-service",
				"local-universal-mesh-service",
			},
		}),
	)
})
