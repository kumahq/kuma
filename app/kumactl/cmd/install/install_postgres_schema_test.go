package install_test

import (
	"bytes"
	"github.com/Kong/kuma/app/kumactl/cmd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"path/filepath"
)

var _ = Describe("kumactl install postgres-schema", func() {
	It("should give the schema postgres", func() {
		// given
		stdout := bytes.Buffer{}
		rootCmd := cmd.DefaultRootCmd()
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"install", "postgres-schema"})

		bytes, err := ioutil.ReadFile(filepath.Join("testdata", "postgres_schema.sql"))
		Expect(err).ToNot(HaveOccurred())

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(stdout.Bytes())).To(Equal(string(bytes)))
	})
})
