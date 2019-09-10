package install_test

import (
	"bytes"
	"github.com/Kong/kuma/app/kumactl/cmd"
	"github.com/Kong/kuma/app/kumactl/pkg/install/data"
	postgres_schema "github.com/Kong/kuma/app/kumactl/pkg/install/universal/control-plane/postgres"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("kumactl install control-plane", func() {
	It("should give the schema postgres", func() {
		// given
		stdout := bytes.Buffer{}
		rootCmd := cmd.DefaultRootCmd()
		rootCmd.SetOut(&stdout)
		rootCmd.SetArgs([]string{"install", "postgres-schema"})

		schemaFile, err := data.ReadFile(postgres_schema.Schema, "resource.sql")
		Expect(err).ToNot(HaveOccurred())

		// when
		err = rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(string(stdout.Bytes())).To(Equal(string(schemaFile)))
	})
})
