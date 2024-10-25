package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
	"github.com/kumahq/kuma/pkg/test/matchers"
)

var _ = Describe("Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	stop := func() {}
	BeforeAll(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(store.NewPaginationStore(resourceStore)))
	})

	AfterAll(func() {
		stop()
	})

	DescribeTable("inspect for policies /meshes/{mesh}/{policyType}/{policyName}/_resources/dataplanes", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/policies/_resources/dataplanes"))

	DescribeTable("inspect for policies /meshes/{mesh}/{serviceType}/{policyName}/_resources/hostnames", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/services/_resources/hostnames"))

	DescribeTable("inspect dataplane rules /meshes/{mesh}/dataplanes/{dpName}/_rules", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_rules"))

	DescribeTable("inspect meshgateway rules /meshes/{mesh}/meshgateways/{gwName}/_rules", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/meshgateways/_rules"))

	DescribeTable("resources CRUD", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/crud"))

	DescribeTable("service insights", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("service-insights"))

	DescribeTable("insights", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("insights"))

	DescribeTable("base_endpoints",
		func(path string) {
			url := fmt.Sprintf("http://%s%s", apiServer.Address(), path)
			res, err := http.Get(url) // nolint:gosec
			Expect(err).ToNot(HaveOccurred())

			Expect(res).To(HaveHTTPStatus(http.StatusOK))
			b, err := io.ReadAll(res.Body)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(b)).To(matchers.MatchGoldenJSON("testdata", "base_endpoints", fmt.Sprintf("%s.golden.json", strings.ReplaceAll(path, "/", ""))))
		},
		Entry(nil, "/_resources"),
		Entry(nil, "/policies"),
		Entry(nil, "/who-am-i"),
	)
}, Ordered)
