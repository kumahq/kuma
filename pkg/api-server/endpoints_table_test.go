package api_server_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/v2/pkg/api-server"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
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

	DescribeTable("inspect for policies /meshes/{mesh}/{serviceType}/{policyName}/_hostnames", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/services/_resources/hostnames/zone"))

	DescribeTable("inspect meshservice dataplanes /meshes/{mesh}/meshservices/{name}/_dataplanes", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/meshservices/_dataplanes"))

	DescribeTable("inspect dataplane rules /meshes/{mesh}/dataplanes/{dpName}/_rules", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_rules"))

	DescribeTable("inspect meshgateway rules /meshes/{mesh}/meshgateways/{gwName}/_rules", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/meshgateways/_rules"))

	DescribeTable("inspect dataplane layout /meshes/{mesh}/dataplanes/{dpName}/_layout", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_layout"))

	DescribeTable("resources CRUD", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/crud"))

	DescribeTable("inspect dataplane policies /meshes/{mesh}/dataplanes/{dpName}/*/_policies", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_policies"))

	DescribeTable("inspect inbound policies /meshes/{mesh}/dataplanes/{dpName}/inbounds/{inboundKri}/_policies", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_policies/inbounds"))

	DescribeTable("inspect outbound policies /meshes/{mesh}/dataplanes/{dpName}/outbounds/{outboundKri}/_policies", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_policies/outbounds"))

	DescribeTable("inspect outbound policies targeting routes /meshes/{mesh}/dataplanes/{dpName}/outbounds/{outboundKri}/_routes/{routeKri}/_policies", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_policies/routes"))

	DescribeTable("inspect dataplane outbound routes  /meshes/{mesh}/dataplanes/{dpName}/outbounds/*/_routes", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("resources/inspect/dataplanes/_routes"))

	DescribeTable("service insights", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("service-insights"))

	DescribeTable("insights", func(inputFile string) {
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("insights"))

	DescribeTable("base_endpoints",
		func(path string) {
			url := fmt.Sprintf("http://%s%s", apiServer.Address(), path)
			res, err := http.Get(url)
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

	Describe("global mode", func() {
		var globalApiServer *api_server.ApiServer
		var globalResourceStore store.ResourceStore
		stopGlobal := func() {}

		BeforeAll(func() {
			globalResourceStore = memory.NewStore()
			globalApiServer, _, stopGlobal = StartApiServer(NewTestApiServerConfigurer().WithGlobal().WithStore(store.NewPaginationStore(globalResourceStore)))
		})

		AfterAll(func() {
			stopGlobal()
		})

		DescribeTable("inspect for services /meshes/{mesh}/{serviceType}/{policyName}/_hostnames", func(inputFile string) {
			apiTest(inputFile, globalApiServer, globalResourceStore)
		}, test.EntriesForFolder("resources/inspect/services/_resources/hostnames/global"))
	})
}, Ordered)
