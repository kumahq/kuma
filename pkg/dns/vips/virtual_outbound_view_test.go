package vips_test

import (
	. "github.com/onsi/ginkgo/v2"
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
	var exampleA = map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("foo"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{"s": "a"}}}}}
	var exampleB = map[vips.HostnameEntry]vips.VirtualOutbound{vips.NewHostEntry("bar"): {Address: "240.0.0.1", Outbounds: []vips.OutboundEntry{{TagSet: map[string]string{"s": "b"}}}}}
	DescribeTable("Update",
		func(tc updateTestCase) {
			// Given
			given, err := vips.NewVirtualOutboundView(tc.given)
			Expect(err).ToNot(HaveOccurred())
			when, err := vips.NewVirtualOutboundView(tc.when)
			Expect(err).ToNot(HaveOccurred())

			// When
			changes, out := given.Update(when)

			// Then
			Expect(tc.thenChanges).To(Equal(changes))
			expected, err := vips.NewVirtualOutboundView(tc.thenVirtualOutbounds)
			Expect(err).ToNot(HaveOccurred())
			Expect(expected.HostnameEntries()).To(Equal(out.HostnameEntries()))
			for _, k := range expected.HostnameEntries() {
				Expect(expected.Get(k)).To(Equal(out.Get(k)))
			}
		},
		Entry("same noop", updateTestCase{
			given:                exampleA,
			when:                 exampleA,
			thenChanges:          []vips.Change{},
			thenVirtualOutbounds: exampleA,
		}),
		Entry("from empty", updateTestCase{
			given:                map[vips.HostnameEntry]vips.VirtualOutbound{},
			when:                 exampleA,
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: exampleA,
		}),
		Entry("to empty", updateTestCase{
			given:                exampleA,
			when:                 map[vips.HostnameEntry]vips.VirtualOutbound{},
			thenChanges:          []vips.Change{{Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{},
		}),
		Entry("add one/remove one", updateTestCase{
			given:                exampleA,
			when:                 exampleB,
			thenChanges:          []vips.Change{{Type: vips.Add, Entry: vips.NewHostEntry("bar")}, {Type: vips.Remove, Entry: vips.NewHostEntry("foo")}},
			thenVirtualOutbounds: exampleB,
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
