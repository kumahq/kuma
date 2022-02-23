package cmd

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
)

var _ = Describe("root", func() {
	It("should be possible to run `kuma-dp` without a sub-command", func() {
		// given
		cmd := NewRootCmd(kuma_cmd.DefaultRunCmdOpts, DefaultRootContext())
		cmd.SetArgs([]string{})
		// when
		err := cmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())
	})
})
