package memory_test

import (
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_store "github.com/Kong/kuma/pkg/test/store"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("MemoryStore", func() {
	test_store.ExecuteStoreTests(memory.NewStore)
})
