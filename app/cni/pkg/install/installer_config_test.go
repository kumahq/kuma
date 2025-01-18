package install

import (
	"context"
	"os"
	"path"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/test/matchers"
)

const expectedKubeconfig = `# Kubeconfig file for kuma CNI plugin.
apiVersion: v1
kind: Config
clusters:
- name: local
  cluster:
    server: https://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:3000
    certificate-authority-data: YWJjCg==
users:
- name: kuma-cni
  user:
    token: token
contexts:
- name: kuma-cni-context
  context:
    cluster: local
    user: kuma-cni
current-context: kuma-cni-context`

var _ = Describe("InstallerConfig", func() {
	Describe("PostProcess", func() {
		It("should use default CNI config when none is found", func() {
			// given
			ic := InstallerConfig{
				MountedCniNetDir: path.Join("testdata", "nonexistent-dir"),
			}

			// when
			err := ic.PostProcess()

			// then
			Expect(err).To(Not(HaveOccurred()))
			Expect(ic.CniConfName).To(Equal(defaultKumaCniConfName))
		})

		It("should find and use the CNI config file if it exists", func() {
			// given
			ic := InstallerConfig{
				MountedCniNetDir: path.Join("testdata", "find-conf-dir"),
			}

			// when
			err := ic.PostProcess()

			// then
			Expect(err).To(Not(HaveOccurred()))
			Expect(ic.CniConfName).To(Equal("10-flannel.conf"))
		})
	})

	Describe("PrepareKubeconfig", func() {
		It("should successfully prepare kubeconfig file", func() {
			// given
			mockServiceAccountPath := filepath.Join("testdata", "prepare-kubeconfig")
			ic := InstallerConfig{
				KubernetesServiceHost:     "localhost",
				KubernetesServicePort:     "3000",
				KubernetesServiceProtocol: "https",
				MountedCniNetDir:          filepath.Join("testdata", "prepare-kubeconfig"),
				KubeconfigName:            "ZZZ-kuma-cni-kubeconfig",
			}

			// when
			err := ic.PrepareKubeconfig(filepath.Join(mockServiceAccountPath, "token"), filepath.Join(mockServiceAccountPath, "ca.crt"))

			// then
			Expect(err).To(Not(HaveOccurred()))
			// and
			kubeconfig, _ := os.ReadFile(filepath.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig"))
			Expect(kubeconfig).To(matchers.MatchGoldenYAML(filepath.Join("testdata", "prepare-kubeconfig", "ZZZ-kuma-cni-kubeconfig.golden")))
		})
	})

	Describe("GenerateKubeconfigTemplate", func() {
		It("should work properly with unescaped IPv6 addresses", func() {
			// given
			ic := InstallerConfig{
				KubernetesServiceProtocol: "https",
				KubernetesServiceHost:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
				KubernetesServicePort:     "3000",
			}

			// when
			result := ic.GenerateKubeconfigTemplate([]byte("token"), []byte("abc\n"))

			// then
			Expect(result).To(Equal(expectedKubeconfig))
		})
	})
})

var _ = Describe("prepareKumaCniConfig", func() {
	It("should successfully prepare chained kuma CNI file", func() {
		// given
		mockServiceAccountPath := filepath.Join("testdata", "prepare-chained-kuma-config")
		ic := InstallerConfig{
			CniNetworkConfig: kumaCniConfigTemplate,
			MountedCniNetDir: filepath.Join("testdata", "prepare-chained-kuma-config"),
			KubeconfigName:   "ZZZ-kuma-cni-kubeconfig",
			CniConfName:      "10-calico.conflist",
			ChainedCniPlugin: true,
		}

		// when
		err := prepareKumaCniConfig(context.Background(), &ic, filepath.Join(mockServiceAccountPath, "token"))

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := os.ReadFile(filepath.Join("testdata", "prepare-chained-kuma-config", "10-calico.conflist"))
		Expect(kubeconfig).To(matchers.MatchGoldenJSON(filepath.Join("testdata", "prepare-chained-kuma-config", "10-calico.conflist.golden")))
	})

	It("should successfully prepare standalone kuma CNI file", func() {
		// given
		mockServiceAccountPath := filepath.Join("testdata", "prepare-standalone-kuma-config")
		ic := InstallerConfig{
			CniNetworkConfig: kumaCniConfigTemplate,
			MountedCniNetDir: filepath.Join("testdata", "prepare-standalone-kuma-config"),
			KubeconfigName:   "ZZZ-kuma-cni-kubeconfig",
			CniConfName:      "kuma-cni.conf",
			ChainedCniPlugin: false,
		}

		// when
		err := prepareKumaCniConfig(context.Background(), &ic, filepath.Join(mockServiceAccountPath, "token"))

		// then
		Expect(err).To(Not(HaveOccurred()))
		// and
		kubeconfig, _ := os.ReadFile(filepath.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf"))
		Expect(kubeconfig).To(matchers.MatchGoldenJSON(filepath.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf.golden")))
	})
})
