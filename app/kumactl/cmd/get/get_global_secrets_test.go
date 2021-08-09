package get_test

import (
	"bytes"
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/wrapperspb"

	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/system"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	. "github.com/kumahq/kuma/pkg/test/matchers"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get global-secrets", func() {

	var sampleSecrets []*system.GlobalSecretResource

	BeforeEach(func() {
		sampleSecrets = []*system.GlobalSecretResource{
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "sec-1",
				},
				Spec: &system_proto.Secret{
					Data: &wrapperspb.BytesValue{
						Value: []byte("test"),
					},
				},
			},
			{
				Meta: &test_model.ResourceMeta{
					Mesh: "default",
					Name: "sec-2",
				},
				Spec: &system_proto.Secret{
					Data: &wrapperspb.BytesValue{
						Value: []byte("test2"),
					},
				},
			},
		}
	})

	Describe("GetGlobalSecretsCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, system.GlobalSecretResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, pt := range sampleSecrets {
				key := core_model.ResourceKey{
					Mesh: pt.Meta.GetMesh(),
					Name: pt.Meta.GetName(),
				}
				err := store.Create(context.Background(), pt, core_store.CreateBy(key))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			goldenFile   string
		}

		DescribeTable("kumactl get secrets -o table|json|yaml",
			func(given testCase) {
				// given
				rootCmd.SetArgs(append([]string{
					"--config-file", filepath.Join("..", "testdata", "sample-kumactl.config.yaml"),
					"get", "global-secrets"}, given.outputFormat))

				// when
				err := rootCmd.Execute()

				// then
				Expect(err).ToNot(HaveOccurred())
				Expect(buf.String()).To(MatchGoldenEqual(filepath.Join("testdata", given.goldenFile)))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-global-secrets.golden.txt",
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-global-secrets.golden.txt",
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-global-secrets.golden.json",
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-global-secrets.golden.yaml",
			}),
		)
	})
})
