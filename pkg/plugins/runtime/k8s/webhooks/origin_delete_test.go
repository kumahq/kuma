package webhooks_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/core"
)

// Regression coverage for https://github.com/kumahq/kuma/issues/10260: a
// non-privileged user should not be able to delete a policy that was synced
// in from elsewhere, because the delete only ever appears to succeed -- the
// next KDS sync just re-creates the resource.
//
// cli-user_delete_zone-mtp.* (in testdata/validation) already covers the
// Global-CP/system-namespace half of this via the pre-existing label-value
// check in pkg/core/resources/labels/registry.go, which only applies to
// plugin-originated resources deleted from a namespace that equals the
// configured system namespace. This spec covers the gap that check doesn't
// reach: a federated Zone deleting a Global-originated policy that lives in
// an ordinary application namespace, which is how policies are normally
// applied on a Zone (Zone, unlike Global, does not require policies to live
// in the system namespace).
var _ = Describe("Delete of a synced-in resource (issue #10260)", func() {
	const fixture = "testdata/validation-origin-delete/global-origin-mtp-in-app-namespace.input.yaml"

	It("should deny deleting a Global-originated policy from a federated Zone", func() {
		// given a federated Zone and a MeshTrafficPermission labeled kuma.io/origin: global,
		// living in an application namespace (not the system namespace)
		wh := newValidatingWebhook(core.Zone, true)
		req := webhookRequest(fixture)

		// when
		resp := wh.Handle(context.Background(), req)

		// then the delete is rejected, not silently allowed to revert on the next sync
		Expect(resp.Allowed).To(BeFalse())
		Expect(resp.Result).ToNot(BeNil())
		Expect(resp.Result.Message).To(ContainSubstring("originated on the global control plane"))
		Expect(resp.Result.Message).To(ContainSubstring("kuma.io/origin: global"))
	})

	It("should still allow deleting the same resource from a non-federated Zone", func() {
		// given a non-federated Zone, which owns everything in its own store
		wh := newValidatingWebhook(core.Zone, false)
		req := webhookRequest(fixture)

		// when
		resp := wh.Handle(context.Background(), req)

		// then
		Expect(resp.Allowed).To(BeTrue())
	})
})
