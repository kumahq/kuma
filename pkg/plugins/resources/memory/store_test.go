package memory_test

import (
	"github.com/Kong/kuma/pkg/core/resources/store"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	test_store "github.com/Kong/kuma/pkg/test/store"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("MemoryStore", func() {
	createStore := func() store.ResourceStore {
		return memory.NewStore()
	}

	test_store.ExecuteStoreTests(createStore)
})
