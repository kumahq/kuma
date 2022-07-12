package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("kubeconfigTemplate", func() {
	It("should work properly with IPv6 addresses", func() {
		// given
		expected := `# Kubeconfig file for kuma CNI plugin.
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

		// when
		result := kubeconfigTemplate("https", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", "3000", "token", "YWJjCg==")

		// then
		Expect(result).To(Equal(expected))

		// when
		resultWithBrackets := kubeconfigTemplate("https", "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]", "3000", "token", "YWJjCg==")

		// then
		Expect(resultWithBrackets).To(Equal(expected))
	})
})
