package install

import (
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("testTransformJsonConfig", func() {
	It("should properly manipulate CNI conflist file", func() {
		// given
		calicoConfig, _ := os.ReadFile(path.Join("testdata", "10-calico.conflist"))
		expectedConfig := path.Join("testdata", "10-calico-cni-injected.conflist")

		// when
		result, err := transformJsonConfig(kumaCniConfig, calicoConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(matchers.MatchGoldenJSON(expectedConfig))
	})

	It("should properly manipulate CNI conf file", func() {
		// given
		calicoConfig, _ := os.ReadFile(path.Join("testdata", "10-flannel.conf"))
		expectedConfig := path.Join("testdata", "10-flannel-cni-injected.conf")

		// when
		result, _ := transformJsonConfig(kumaCniConfig, calicoConfig)

		// then
		Expect(result).To(matchers.MatchGoldenJSON(expectedConfig))
	})
})

var _ = Describe("revertConfig", func() {
	It("should properly revert CNI conflist", func() {
		// given
		changedConfig, _ := os.ReadFile(path.Join("testdata", "10-calico-cni-injected.conflist"))
		originalConfig := path.Join("testdata", "10-calico.conflist")

		// when
		result, err := revertConfigContents(changedConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(matchers.MatchGoldenJSON(originalConfig))
	})

	It("should properly revert CNI conf", func() {
		// given
		changedConfig, _ := os.ReadFile(path.Join("testdata", "10-flannel-cni-injected.conf"))
		originalConfig := path.Join("testdata", "10-flannel-clean.conf")

		// when
		result, err := revertConfigContents(changedConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(matchers.MatchGoldenJSON(originalConfig))
	})
})
