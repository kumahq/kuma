package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const expected = `# Kubeconfig file for kuma CNI plugin.
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

var _ = Describe("kubeconfigTemplate", func() {
	It("should work properly with unescaped IPv6 addresses", func() {
		// given
		ic := InstallerConfig{
			KubernetesServiceProtocol: "https",
			KubernetesServiceHost:     "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			KubernetesServicePort:     "3000",
		}

		// when
		result := kubeconfigTemplate(&ic, "token", "YWJjCg==")

		// then
		Expect(result).To(Equal(expected))

	})

	It("should work properly with escaped IPv6 addresses", func() {
		// given
		ic := InstallerConfig{
			KubernetesServiceProtocol: "https",
			KubernetesServiceHost:     "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]",
			KubernetesServicePort:     "3000",
		}

		// when
		resultWithBrackets := kubeconfigTemplate(&ic, "token", "YWJjCg==")

		// then
		Expect(resultWithBrackets).To(Equal(expected))
	})
})
