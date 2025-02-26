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
				KubernetesServiceHost:             "localhost",
				KubernetesServicePort:             "3000",
				KubernetesServiceProtocol:         "https",
				KubernetesCaFile:                  filepath.Join(mockServiceAccountPath, "ca.crt"),
				KubernetesServiceAccountTokenPath: filepath.Join(mockServiceAccountPath, "token"),
				MountedCniNetDir:                  filepath.Join("testdata", "prepare-kubeconfig"),
				KubeconfigName:                    "ZZZ-kuma-cni-kubeconfig",
			}

			// when
			err := ic.PrepareKubeconfig()

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

	Describe("PrepareKumaCniConfig", func() {
		It("should successfully prepare chained kuma CNI file", func() {
			// given
			mockServiceAccountPath := filepath.Join("testdata", "prepare-chained-kuma-config")
			ic := InstallerConfig{
				CniNetworkConfig:                  kumaCniConfigTemplate,
				MountedCniNetDir:                  mockServiceAccountPath,
				HostCniNetDir:                     "/foo/bar",
				KubeconfigName:                    "ZZZ-kuma-cni-kubeconfig",
				KubernetesServiceAccountTokenPath: filepath.Join(mockServiceAccountPath, "token"),
				CniConfName:                       "10-calico.conflist",
				ChainedCniPlugin:                  true,
			}

			// when
			err := ic.PrepareKumaCniConfig(context.Background())

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
				CniNetworkConfig:                  kumaCniConfigTemplate,
				MountedCniNetDir:                  mockServiceAccountPath,
				HostCniNetDir:                     "/etc/cni/net.d",
				KubeconfigName:                    "ZZZ-kuma-cni-kubeconfig",
				KubernetesServiceAccountTokenPath: filepath.Join(mockServiceAccountPath, "token"),
				CniConfName:                       "kuma-cni.conf",
				ChainedCniPlugin:                  false,
			}

			// when
			err := ic.PrepareKumaCniConfig(context.Background())

			// then
			Expect(err).To(Not(HaveOccurred()))
			// and
			kubeconfig, _ := os.ReadFile(filepath.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf"))
			Expect(kubeconfig).To(matchers.MatchGoldenJSON(filepath.Join("testdata", "prepare-standalone-kuma-config", "kuma-cni.conf.golden")))
		})
	})

	Context("CheckInstall", func() {
		It("should not return an error when a file is a conflist file with kuma-cni installed", func() {
			// given
			ic := InstallerConfig{
				MountedCniNetDir: "testdata",
				CniConfName:      "10-flannel-cni-injected.conf",
				ChainedCniPlugin: true,
			}

			// when
			err := ic.CheckInstall()

			// then
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should not return an error when a file is a conf file with kuma-cni", func() {
			// given
			ic := InstallerConfig{
				MountedCniNetDir: "testdata",
				CniConfName:      "10-kuma-cni.conf",
				ChainedCniPlugin: false,
			}

			// when
			err := ic.CheckInstall()

			// then
			Expect(err).To(Not(HaveOccurred()))
		})

		It("should return an error when a file does not have kuma-cni installed", func() {
			// given
			ic := InstallerConfig{
				MountedCniNetDir: "testdata",
				CniConfName:      "10-flannel.conf",
				ChainedCniPlugin: false,
			}

			// when
			err := ic.CheckInstall()

			// then
			Expect(err).To(HaveOccurred())
		})
	})
})
