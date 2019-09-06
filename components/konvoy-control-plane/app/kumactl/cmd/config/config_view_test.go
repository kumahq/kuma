package config_test

import (
	"bytes"
	"github.com/Kong/konvoy/components/konvoy-control-plane/app/kumactl/cmd"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("konvoy config view", func() {

	It("should display configuration from a given file", func() {
		// setup
		rootCmd := cmd.DefaultRootCmd()
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "config-view.config.yaml"),
			"config", "view"})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := ioutil.ReadFile(filepath.Join("testdata", "config-view.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(buf.String()).To(MatchYAML(expected))
	})
})
