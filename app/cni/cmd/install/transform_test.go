package main

import (
	"io/ioutil"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const kumaCniConfig = `{
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

var _ = Describe("testTransformJsonConfig", func() {
	It("should properly manipulate CNI conflist file", func() {
		// given
		calicoConfig, _ := ioutil.ReadFile("testdata/10-calico.conflist")
		expectedConfig, _ := ioutil.ReadFile("testdata/10-calico.conflist.golden")

		// when
		result, err := transformJsonConfig(kumaCniConfig, calicoConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(MatchJSON(expectedConfig))
	})

	It("should properly manipulate CNI conf file", func() {
		// given
		calicoConfig, _ := ioutil.ReadFile("testdata/10-flannel.conf")
		expectedConfig, _ := ioutil.ReadFile("testdata/10-flannel.conf.golden")

		// when
		result, _ := transformJsonConfig(kumaCniConfig, calicoConfig)

		// then
		Expect(result).To(MatchJSON(expectedConfig))
	})
})

var _ = Describe("revertConfig", func() {
	It("should properly revert CNI conflist", func() {
		// given
		changedConfig, _ := ioutil.ReadFile("testdata/10-calico.conflist.golden")
		originalConfig, _ := ioutil.ReadFile("testdata/10-calico.conflist")

		// when
		result, err := revertConfigContents(changedConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(MatchJSON(originalConfig))
	})

	It("should properly revert CNI conf", func() {
		// given
		changedConfig, _ := ioutil.ReadFile("testdata/10-flannel.conf.golden")
		originalConfig, _ := ioutil.ReadFile("testdata/10-flannel-clean.conf.golden")

		// when
		result, err := revertConfigContents(changedConfig)

		// then
		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(MatchJSON(originalConfig))
	})
})
