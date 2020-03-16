package install_test

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Kong/kuma/app/kumactl/cmd"
)

var _ = Describe("kumactl install database-schema", func() {
	It("should give the schema postgres", func() {
		// given
		stdout := bytes.Buffer{}
		rootCmd := cmd.DefaultRootCmd()
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"install", "database-schema", "--target=postgres"})

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "postgres_schema.golden.sql"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(stdout.String()).To(Equal(string(expected)))
	})
})
