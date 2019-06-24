package memory_test

import (
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/store"
	"github.com/Kong/konvoy/components/konvoy-control-plane/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("MemoryStore", func() {
	var c store.ResourceStore

	BeforeEach(func() {
		c = store.NewStrictResourceStore(memory.NewStore())
	})

	store.ExecuteStoreTests(&c)
})
