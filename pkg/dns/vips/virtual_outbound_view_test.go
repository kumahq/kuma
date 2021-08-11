package vips_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/dns/vips"
)

var _ = Describe("Virtual outbound", func() {
	type updateTestCase struct {
		given                map[vips.HostnameEntry]vips.VirtualOutbound
		when                 map[vips.HostnameEntry]vips.VirtualOutbound
		thenChanges          []vips.Change
		thenVirtualOutbounds map[vips.HostnameEntry]vips.VirtualOutbound
	}
	DescribeTable("Update",
		func(tc updateTestCase) {
			changes, out := vips.NewVirtualOutboundView(tc.given).Update(vips.NewVirtualOutboundView(tc.when))

			Expect(tc.thenChanges).To(Equal(changes))
			expected := vips.NewVirtualOutboundView(tc.thenVirtualOutbounds)
			Expect(expected.HostnameEntries()).To(Equal(out.HostnameEntries()))
			for _, k := range expected.HostnameEntries() {
				Expect(expected.Get(k)).To(Equal(out.Get(k)))
			}
		},
		Entry("same noop", updateTestCase{
			given:                map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
		}),
		Entry("from empty", updateTestCase{
			given:                map[vips.HostnameEntry]vips.VirtualOutbound{},
			when:                 map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
		}),
		Entry("to empty", updateTestCase{
			given:                map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.HostnameEntry]vips.VirtualOutbound{},
			thenChanges:          []vips.Change{{Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{},
		}),
		Entry("add one/remove one", updateTestCase{
			given:                map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1"}},
			when:                 map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("bar"): {Address: "240.0.0.1"}},
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("bar")}, {Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("bar"): {Address: "240.0.0.1"}},
		}),
		Entry("add extra outbound", updateTestCase{
			given: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"}}}},
			when: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{
				{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"},
				{Port: 81, TagSet: map[string]string{"foo": "baz"}, Origin: "my-policy2"},
			}}},
			thenChanges: []vips.Change{{Type: vips.Modify, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{
				{Port: 80, TagSet: map[string]string{"foo": "bar"}, Origin: "my-policy"},
				{Port: 81, TagSet: map[string]string{"foo": "baz"}, Origin: "my-policy2"},
			}}},
		}),
	)
})
