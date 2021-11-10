package get_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	gomega_types "github.com/onsi/gomega/types"
	"github.com/spf13/cobra"

	"github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/app/kumactl/cmd"
	"github.com/kumahq/kuma/pkg/core/resources/apis/mesh"
	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
	core_store "github.com/kumahq/kuma/pkg/core/resources/store"
	memory_resources "github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_kumactl "github.com/kumahq/kuma/pkg/test/kumactl"
	test_model "github.com/kumahq/kuma/pkg/test/resources/model"
)

var _ = Describe("kumactl get rate-limits", func() {

	rateLimitResources := []*mesh.RateLimitResource{
		{
			Spec: &v1alpha1.RateLimit{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web1",
							"version": "1.0",
						},
					},
				},
				Destinations: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "backend1",
							"env":     "dev",
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "web1-to-backend1",
			},
		},
		{
			Spec: &v1alpha1.RateLimit{
				Sources: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "web2",
							"version": "1.0",
						},
					},
				},
				Destinations: []*v1alpha1.Selector{
					{
						Match: map[string]string{
							"service": "backend2",
							"env":     "dev",
						},
					},
				},
			},
			Meta: &test_model.ResourceMeta{
				Mesh: "default",
				Name: "web2-to-backend2",
			},
		},
	}

	Describe("GetRateLimitsCmd", func() {

		var rootCmd *cobra.Command
		var buf *bytes.Buffer
		var store core_store.ResourceStore
		rootTime, _ := time.Parse(time.RFC3339, "2008-04-27T16:05:36.995Z")
		BeforeEach(func() {
			// setup
			store = core_store.NewPaginationStore(memory_resources.NewStore())

			rootCtx, err := test_kumactl.MakeRootContext(rootTime, store, mesh.RateLimitResourceTypeDescriptor)
			Expect(err).ToNot(HaveOccurred())

			for _, ds := range rateLimitResources {
				err := store.Create(context.Background(), ds, core_store.CreateBy(core_model.MetaToResourceKey(ds.GetMeta())))
				Expect(err).ToNot(HaveOccurred())
			}

			rootCmd = cmd.NewRootCmd(rootCtx)
			buf = &bytes.Buffer{}
			rootCmd.SetOut(buf)
		})

		type testCase struct {
			outputFormat string
			pagination   string
			goldenFile   string
			matcher      func(interface{}) gomega_types.GomegaMatcher
		}

		DescribeTable("kumactl get rate-limits -o table|json|yaml",
			func(given testCase) {
				// when
				Expect(
					ExecuteRootCommand(rootCmd, "rate-limits", given.outputFormat, given.pagination),
				).To(Succeed())

				// when
				expected, err := ioutil.ReadFile(filepath.Join("testdata", given.goldenFile))
				// then
				Expect(err).ToNot(HaveOccurred())
				// and
				Expect(buf.String()).To(given.matcher(expected))
			},
			Entry("should support Table output by default", testCase{
				outputFormat: "",
				goldenFile:   "get-rate-limits.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support Table output explicitly", testCase{
				outputFormat: "-otable",
				goldenFile:   "get-rate-limits.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support pagination", testCase{
				outputFormat: "-otable",
				pagination:   "--size=1",
				goldenFile:   "get-rate-limits.pagination.golden.txt",
				matcher: func(expected interface{}) gomega_types.GomegaMatcher {
					return WithTransform(strings.TrimSpace, Equal(strings.TrimSpace(string(expected.([]byte)))))
				},
			}),
			Entry("should support JSON output", testCase{
				outputFormat: "-ojson",
				goldenFile:   "get-rate-limits.golden.json",
				matcher:      MatchJSON,
			}),
			Entry("should support YAML output", testCase{
				outputFormat: "-oyaml",
				goldenFile:   "get-rate-limits.golden.yaml",
				matcher:      MatchYAML,
			}),
		)
	})

})
