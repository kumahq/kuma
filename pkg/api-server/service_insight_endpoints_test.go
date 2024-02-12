package api_server_test

import (
	. "github.com/onsi/ginkgo/v2"

	api_server "github.com/kumahq/kuma/pkg/api-server"
	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/pkg/test"
)

var _ = Describe("ServiceInsight Endpoints", func() {
	var apiServer *api_server.ApiServer
	var resourceStore store.ResourceStore
	stop := func() {}
	BeforeAll(func() {
		resourceStore = memory.NewStore()
		apiServer, _, stop = StartApiServer(NewTestApiServerConfigurer().WithStore(resourceStore))
	})

	AfterAll(func() {
		stop()
	})

	DescribeTable("table tests", func(inputFile string) {
		memory.ClearStore(resourceStore)
		apiTest(inputFile, apiServer, resourceStore)
	}, test.EntriesForFolder("service-insights"))
}, Ordered)
