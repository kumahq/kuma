package config_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/kumahq/kuma/pkg/util/test"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kumactl config control-planes list", func() {

	It("should display Control Planes from a given configuration file", func() {
		// setup
		rootCmd := test.DefaultTestingRootCmd()
		buf := &bytes.Buffer{}
		rootCmd.SetOut(buf)

		// given
		rootCmd.SetArgs([]string{
			"--config-file", filepath.Join("testdata", "config-control-planes-list.config.yaml"),
			"config", "control-planes", "list"})

		// when
		err := rootCmd.Execute()
		// then
		Expect(err).ToNot(HaveOccurred())

		// when
		expected, err := ioutil.ReadFile(filepath.Join("testdata", "config-control-planes-list.golden.txt"))
		// then
		Expect(err).ToNot(HaveOccurred())
		// and
		Expect(strings.TrimSpace(buf.String())).To(Equal(strings.TrimSpace(string(expected))))
	})
})
