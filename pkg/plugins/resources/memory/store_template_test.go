package memory_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/kumahq/kuma/pkg/plugins/resources/memory"
	test_store "github.com/kumahq/kuma/pkg/test/store"
)

var _ = Describe("MemoryStore template", func() {
	test_store.ExecuteStoreTests(memory.NewStore)
	test_store.ExecuteOwnerTests(memory.NewStore)
})
