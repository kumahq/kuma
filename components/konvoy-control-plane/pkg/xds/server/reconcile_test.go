package server

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
)

var _ = Describe("Reconcile", func() {
	Describe("basicSnapshotGenerator", func() {

		generator := basicSnapshotGenerator{}

		It("should support Nodes without metadata", func() {
			// given
			node := &core.Node{
				Id:      "side-car",
				Cluster: "example",
			}

			// when
			ss := generator.NewSnapshot(node)

			// then
			Expect(ss).ToNot(BeNil())
		})
	})
})
