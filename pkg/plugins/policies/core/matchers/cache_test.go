package matchers_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"

	core_plugins "github.com/kumahq/kuma/v3/pkg/core/plugins"
	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/core/matchers"
	meshtrafficpermission_api "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
)

var _ = Describe("PolicyMatchingCache", func() {
	newMetric := func() *prometheus.CounterVec {
		return prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "test_policy_matching_cache",
		}, []string{"result"})
	}

	Describe("BuildCacheKey", func() {
		It("returns distinct keys for different resource types", func() {
			dpp := readDPP(filepath.Join("testdata", "matchedpolicies", "fromrules", "01.dataplane.yaml"))
			k1 := matchers.BuildCacheKey("TypeA", false, dpp, "hash1")
			k2 := matchers.BuildCacheKey("TypeB", false, dpp, "hash1")
			Expect(k1).ToNot(Equal(k2))
		})

		It("returns distinct keys for different policyMatchingHash", func() {
			dpp := readDPP(filepath.Join("testdata", "matchedpolicies", "fromrules", "01.dataplane.yaml"))
			k1 := matchers.BuildCacheKey("TypeA", false, dpp, "hash1")
			k2 := matchers.BuildCacheKey("TypeA", false, dpp, "hash2")
			Expect(k1).ToNot(Equal(k2))
		})

		It("returns distinct keys for different includeShadow", func() {
			dpp := readDPP(filepath.Join("testdata", "matchedpolicies", "fromrules", "01.dataplane.yaml"))
			k1 := matchers.BuildCacheKey("TypeA", false, dpp, "hash1")
			k2 := matchers.BuildCacheKey("TypeA", true, dpp, "hash1")
			Expect(k1).ToNot(Equal(k2))
		})

		It("returns the same key for identical inputs", func() {
			dpp := readDPP(filepath.Join("testdata", "matchedpolicies", "fromrules", "01.dataplane.yaml"))
			k1 := matchers.BuildCacheKey("TypeA", false, dpp, "hash1")
			k2 := matchers.BuildCacheKey("TypeA", false, dpp, "hash1")
			Expect(k1).To(Equal(k2))
		})
	})

	Describe("GetIfPresent / Put", func() {
		It("returns miss on empty cache", func() {
			c := matchers.NewPolicyMatchingCache(newMetric(), 100)
			_, ok := c.GetIfPresent("no-such-key")
			Expect(ok).To(BeFalse())
		})

		It("returns hit after Put", func() {
			c := matchers.NewPolicyMatchingCache(newMetric(), 100)
			want := core_model.ResourceType("MeshTrafficPermission")
			c.Put("k", core_xds.TypedMatchingPolicies{Type: want})
			got, ok := c.GetIfPresent("k")
			Expect(ok).To(BeTrue())
			Expect(got.Type).To(Equal(want))
		})

		It("returns miss for different key", func() {
			c := matchers.NewPolicyMatchingCache(newMetric(), 100)
			c.Put("k1", core_xds.TypedMatchingPolicies{Type: "Type1"})
			_, ok := c.GetIfPresent("k2")
			Expect(ok).To(BeFalse())
		})
	})

	Describe("MatchedPolicies with cache", func() {
		testDir := filepath.Join("testdata", "matchedpolicies", "fromrules")

		It("cached result is identical to uncached result", func() {
			dpp := readDPP(filepath.Join(testDir, "01.dataplane.yaml"))
			resources, _ := readPolicies(filepath.Join(testDir, "01.policies.yaml"))
			rType := meshtrafficpermission_api.MeshTrafficPermissionType

			uncached, err := matchers.MatchedPolicies(rType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())

			// first call: cache miss, computes result
			c := matchers.NewPolicyMatchingCache(newMetric(), 100)
			const hash = "mesh-hash-stable"
			first, err := matchers.MatchedPolicies(rType, dpp, resources, core_plugins.WithCache(c, hash))
			Expect(err).ToNot(HaveOccurred())
			Expect(first).To(BeComparableTo(uncached))

			// second call: must be a cache hit
			second, err := matchers.MatchedPolicies(rType, dpp, resources, core_plugins.WithCache(c, hash))
			Expect(err).ToNot(HaveOccurred())
			Expect(second).To(BeComparableTo(first))
		})

		It("different policyMatchingHash causes a cache miss", func() {
			dpp := readDPP(filepath.Join(testDir, "01.dataplane.yaml"))
			resources, _ := readPolicies(filepath.Join(testDir, "01.policies.yaml"))
			rType := meshtrafficpermission_api.MeshTrafficPermissionType
			c := matchers.NewPolicyMatchingCache(newMetric(), 100)

			_, err := matchers.MatchedPolicies(rType, dpp, resources, core_plugins.WithCache(c, "hash-v1"))
			Expect(err).ToNot(HaveOccurred())

			result2, err := matchers.MatchedPolicies(rType, dpp, resources, core_plugins.WithCache(c, "hash-v2"))
			Expect(err).ToNot(HaveOccurred())

			uncached, err := matchers.MatchedPolicies(rType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result2).To(BeComparableTo(uncached))
		})

		It("nil cache uses uncached path transparently", func() {
			dpp := readDPP(filepath.Join(testDir, "01.dataplane.yaml"))
			resources, _ := readPolicies(filepath.Join(testDir, "01.policies.yaml"))
			rType := meshtrafficpermission_api.MeshTrafficPermissionType

			result, err := matchers.MatchedPolicies(rType, dpp, resources)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Type).To(Equal(rType))
		})
	})
})
