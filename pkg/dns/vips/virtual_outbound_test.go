package vips_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns/vips"
)

var _ = Describe("Virtual outbound", func() {
	type updateTestCase struct {
		given                map[vips.Entry]vips.VirtualOutbound
		when                 map[vips.Entry]vips.VirtualOutbound
		thenChanges          []vips.Change
		thenVirtualOutbounds map[vips.Entry]vips.VirtualOutbound
	}
	DescribeTable("Update",
		func(tc updateTestCase) {
			changes, out := vips.NewVirtualOutboundView(tc.given).Update(vips.NewVirtualOutboundView(tc.when))

			Expect(tc.thenChanges).To(Equal(changes))
			expected := vips.NewVirtualOutboundView(tc.thenVirtualOutbounds)
			Expect(expected.Keys()).To(Equal(out.Keys()))
			for _, k := range expected.Keys() {
				Expect(expected.Get(k)).To(Equal(out.Get(k)))
			}
		},
		Entry("same noop", updateTestCase{
			given:                map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{},
			thenVirtualOutbounds: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
		}),
		Entry("from empty", updateTestCase{
			given:                map[vips.Entry]vips.VirtualOutbound{},
			when:                 map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
		}),
		Entry("to empty", updateTestCase{
			given:                map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.Entry]vips.VirtualOutbound{},
			thenChanges:          []vips.Change{{Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.Entry]vips.VirtualOutbound{},
		}),
		Entry("add one/remove one", updateTestCase{
			given:                map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("bar"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("bar")}, {Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("bar"): {Address: "240.0.0.1"}},
		}),
		Entry("add extra outbound", updateTestCase{
			given: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.MeshOutbound{{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"}}}},
			when: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.MeshOutbound{
				{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"},
				{Port: 81, TagSet: map[string]string{"foo": "baz"}, Origin: "my-policy2"},
			}}},
			thenChanges: []vips.Change{{Type: vips.Modify, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.Entry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.MeshOutbound{
				{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"},
				{Port: 81, TagSet: map[string]string{"foo": "baz"}, Origin: "my-policy2"},
			}}},
		}),
	)
})
