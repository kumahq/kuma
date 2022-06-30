package install

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"io/ioutil"
)

var _ = Describe("TestTransformJsonConfig", func () {
	It("should properly manipulate CNI config", func() {
		// given
		kumaCniConfig := `{
		  "type": "kuma-cni",
		  "log_level": "info",
		  "kubernetes": {
			"kubeconfig": "/etc/cni/net.d/ZZZ-kuma-cni-kubeconfig",
			"cni_bin_dir": "/opt/cni/bin",
			"exclude_namespaces": [
			  "kuma-system"
			]
		  }
		}`
		calicoConfig, _ := ioutil.ReadFile("data/given/10-calico.conflist")
		expectedConfig, _ := ioutil.ReadFile("data/expected/10-calico.conflist")

		// when
		result, _ := transformJsonConfig(kumaCniConfig, calicoConfig)

		// then
		Expect(result).To(MatchJSON(expectedConfig))
	})
})
