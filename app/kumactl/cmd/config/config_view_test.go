package config_test

import (
	"bytes"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/util/test"
)

var _ = Describe("kumactl config view", func() {

	It("should display configuration from a given file", func() {
		// setup
		rootCmd := test.DefaultTestingRootCmd()
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
		expected, err := os.ReadFile(filepath.Join("testdata", "config-view.golden.yaml"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(buf.String()).To(MatchYAML(expected))
	})
})
