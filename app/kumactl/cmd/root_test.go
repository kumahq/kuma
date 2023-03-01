package cmd_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/app/kumactl/pkg/config"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
)

var _ = Describe("kumactl root cmd", func() {
	var backupDefaultConfigFile string

	BeforeEach(func() {
		file, err := os.CreateTemp("", "")
		Expect(err).To(Succeed())
		// we have to remove file. Config is created only if file does not exist already
		Expect(os.Remove(file.Name())).To(Succeed())
		backupDefaultConfigFile = config.DefaultConfigFile
		config.DefaultConfigFile = file.Name()
	})

	AfterEach(func() {
		config.DefaultConfigFile = backupDefaultConfigFile
	})

	It("should create default config at startup", func() {
		// given
		rootCtx := test_kumactl.MakeMinimalRootContext()
		rootCmd := cmd.NewRootCmd(rootCtx)

		// when
		rootCmd.SetArgs([]string{"version"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		// and default config is created
		expected := `
contexts:
- controlPlane: local
  name: local
controlPlanes:
- coordinates:
    apiServer:
      url: http://localhost:5681
  name: local
currentContext: local
`
		bytes, err := os.ReadFile(config.DefaultConfigFile)
		Expect(err).ToNot(HaveOccurred())
		Expect(bytes).To(MatchYAML(expected))
	})

	It("shouldn't create config file when --no-config flag is set", func() {
		// given
		rootCtx := test_kumactl.MakeMinimalRootContext()
		rootCmd := cmd.NewRootCmd(rootCtx)

		// when
		rootCmd.SetArgs([]string{"version", "--no-config"})
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())

		_, err = os.Stat(config.DefaultConfigFile)
		Expect(err).To(HaveOccurred())
		Expect(os.IsNotExist(err)).To(BeTrue())
	})
})
