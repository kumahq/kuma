package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/pkg/cache"
)

var _ = Describe("Reconcile", func() {
	Describe("reconciler", func() {

		var stopCh chan struct{}
		var hasher hasher
		var logger logger

		BeforeEach(func() {
			stopCh = make(chan struct{})
		})

		AfterEach(func() {
			close(stopCh)
		})

		It("should generate a Snaphot per Envoy Node", func() {
			// setup
			store := cache.NewSnapshotCache(true, hasher, logger)
			nodes := make(chan *core.Node)
			r := &reconciler{nodes, &basicSnapshotGenerator{}, &simpleSnapshotCacher{hasher, store}}

			// given
			node := &core.Node{
				Id:      "side-car",
				Cluster: "example",
			}

			// when
			go r.Start(stopCh)
			nodes <- node

			// then
			Eventually(func() bool {
				_, err := store.GetSnapshot(hasher.ID(node))
				return err == nil
			}, "1s", "1ms").Should(BeTrue())

			// when
			close(nodes)

			// then
			// nothing bad should happen
		})
	})
})
