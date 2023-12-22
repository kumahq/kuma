package export_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	"github.com/kumahq/kuma/pkg/test/matchers"
	"github.com/kumahq/kuma/pkg/test/resources/samples"
)

var _ = Describe("kumactl export", func() {
	var rootCmd *cobra.Command
	var store core_store.ResourceStore
	var buf *bytes.Buffer

	rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")

	BeforeEach(func() {
		store = core_store.NewPaginationStore(memory_resources.NewStore())
		rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, mesh.MeshResourceTypeDescriptor, system.SecretResourceTypeDescriptor, system.GlobalSecretResourceTypeDescriptor)
		Expect(err).ToNot(HaveOccurred())

		rootCmd = cmd.NewRootCmd(rootCtx)
		buf = &bytes.Buffer{}
		rootCmd.SetOut(buf)
	})

	It("should export resources in universal format", func() {
		// given
		resources := []model.Resource{
			samples.MeshDefault(),
			samples.SampleSigningKeyGlobalSecret(),
			samples.SampleSigningKeySecret(),
			samples.MeshDefaultBuilder().WithName("another-mesh").Build(),
			samples.SampleSigningKeySecretBuilder().WithMesh("another-mesh").Build(),
		}
		for _, res := range resources {
			err := store.Create(context.Background(), res, core_store.CreateByKey(res.GetMeta().GetName(), res.GetMeta().GetMesh()))
			Expect(err).ToNot(HaveOccurred())
		}

		args := []string{
			"--config-file",
			filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
			"export",
		}
		rootCmd.SetArgs(args)

		// when
		err := rootCmd.Execute()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(matchers.MatchGoldenEqual("testdata", "export.golden.yaml"))
	})

	type testCase struct {
		args []string
		err  string
	}
	DescribeTable("should fail on invalid resource",
		func(given testCase) {
			// given
			args := []string{
				"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
				"export",
			}
			args = append(args, given.args...)
			rootCmd.SetArgs(args)

			// when
			err := rootCmd.Execute()

			// then
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(given.err))
		},
		Entry("invalid profile", testCase{
			args: []string{"--profile", "something"},
			err:  "invalid profile",
		}),
		Entry("invalid format", testCase{
			args: []string{"--format", "something"},
			err:  "invalid format",
		}),
	)
})
