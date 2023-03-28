package memory_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/core/resources/store"
	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("MemoryStore template", func() {
	test_store.ExecuteStoreTests(func(string) store.ResourceStore { return memory.NewStore() }, "memory")
	test_store.ExecuteOwnerTests(func(string) store.ResourceStore { return memory.NewStore() }, "memory")
})
