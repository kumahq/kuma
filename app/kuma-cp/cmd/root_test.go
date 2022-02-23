package cmd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("root", func() {
	It("should be possible to run `kuma-cp` without a sub-command", func() {
		// given
		cmd := newRootCmd()
		cmd.SetArgs([]string{})
		// when
		err := cmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
