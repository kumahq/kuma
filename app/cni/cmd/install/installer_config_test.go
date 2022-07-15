package main

import (
	"io/ioutil"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("findCniConfFile", func() {
	It("should find conf file in a flat dir", func() {
		// given
		dir := path.Join("testdata", "find-conf-dir")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(Equal("10-flannel.conf"))
	})

	It("should find conflist file in a dir", func() {
		// given
		dir := path.Join("testdata", "find-conflist-dir")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(Equal("10-calico.conflist"))
	})

	It("should not find conf file in a nested dir", func() {
		// given
		dir := path.Join("testdata", "find-conf-dir-nested")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(HaveOccurred())
		Expect(result).To(Equal(""))
	})
})

var _ = Describe("prepareKubeconfig", func() {
	It("should find conf file in a flat dir", func() {
		// given
		mockServiceAccountPath := path.Join("testdata", "prepare-kubeconfig")
		ic := InstallerConfig{
			KubernetesServiceHost: "localhost",
			KubernetesServicePort: "3000",
			MountedCniNetDir:      path.Join("testdata", "prepare-kubeconfig"),
			KubeconfigName:        "ZZZ-kuma-cni-kubeconfig",
		}

		// when
		err := prepareKubeconfig(&ic, mockServiceAccountPath)

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := ioutil.ReadFile(path.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig"))
		Expect(kubeconfig).To(matchers.MatchGoldenEqual(path.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig.golden")))
	})
})
