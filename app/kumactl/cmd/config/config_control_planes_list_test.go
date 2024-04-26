package config_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/pkg/test"
)

var _ = Describe("kumactl config control-planes list", func() {
	It("should display Control Planes from a given configuration file", func() {
		// setup
		buf, _, rootCmd := test.DefaultTestingRootCmd(
			"--config-file", filepath.Join("testdata", "config-control-planes-list.config.yaml"),
			"config", "control-planes", "list",
		)

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := os.ReadFile(filepath.Join("testdata", "config-control-planes-list.golden.txt"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(string(expected))))
	})
})
