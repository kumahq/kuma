package install

import (
	"context"
	"os"
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
	It("should successfully prepare kubeconfig file", func() {
		// given
		mockServiceAccountPath := path.Join("testdata", "prepare-kubeconfig")
		ic := InstallerConfig{
			KubernetesServiceHost:     "localhost",
			KubernetesServicePort:     "3000",
			KubernetesServiceProtocol: "https",
			MountedCniNetDir:          path.Join("testdata", "prepare-kubeconfig"),
			KubeconfigName:            "ZZZ-kuma-cni-kubeconfig",
		}

		// when
		err := prepareKubeconfig(&ic, mockServiceAccountPath + "/token", mockServiceAccountPath + "/ca.crt")

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := os.ReadFile(path.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig"))
		Expect(kubeconfig).To(matchers.MatchGoldenYAML(path.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig.golden")))
	})
})

var _ = Describe("prepareKumaCniConfig", func() {
	It("should successfully prepare chained kuma cni file", func() {
		// given
		mockServiceAccountPath := path.Join("testdata", "prepare-chained-kuma-config")
		ic := InstallerConfig{
			CniNetworkConfig: kumaCniConfigTemplate,
			MountedCniNetDir: path.Join("testdata", "prepare-chained-kuma-config"),
			KubeconfigName:   "ZZZ-kuma-cni-kubeconfig",
			CniConfName:      "10-calico.conflist",
			ChainedCniPlugin: true,
		}

		// when
		err := prepareKumaCniConfig(context.Background(), &ic, mockServiceAccountPath + "/token")

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := os.ReadFile(path.Join("testdata", "prepare-chained-kuma-config", "10-calico.conflist"))
		Expect(kubeconfig).To(matchers.MatchGoldenJSON(path.Join("testdata", "prepare-chained-kuma-config", "10-calico.conflist.golden")))
	})

	It("should successfully prepare standalone kuma cni file", func() {
		// given
		mockServiceAccountPath := path.Join("testdata", "prepare-standalone-kuma-config")
		ic := InstallerConfig{
			CniNetworkConfig: kumaCniConfigTemplate,
			MountedCniNetDir: path.Join("testdata", "prepare-standalone-kuma-config"),
			KubeconfigName:   "ZZZ-kuma-cni-kubeconfig",
			CniConfName:      "kuma-cni.conf",
			ChainedCniPlugin: false,
		}

		// when
		err := prepareKumaCniConfig(context.Background(), &ic, mockServiceAccountPath + "/token")

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := os.ReadFile(path.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf"))
		Expect(kubeconfig).To(matchers.MatchGoldenJSON(path.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf.golden")))
	})
})
