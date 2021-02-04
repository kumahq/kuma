package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("root", func() {
	It("should be possible to run `kuma-dp` without a sub-command", func() {
		// given
		cmd := NewRootCmd(DefaultRootContext())
		cmd.SetArgs([]string{})
		// when
		err := cmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
