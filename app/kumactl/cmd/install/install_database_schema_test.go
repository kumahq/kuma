package install_test

import (
	"bytes"
	"github.com/Kong/kuma/app/kumactl/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
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
