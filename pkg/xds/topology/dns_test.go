package topology_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/xds"
	"github.com/kumahq/kuma/pkg/dns/vips"
	"github.com/kumahq/kuma/pkg/xds/topology"
)

var _ = Describe("VIPOutbounds", func() {
	type outboundTestCase struct {
		whenOutbounds map[vips.HostnameEntry]vips.VirtualOutbound
		thenVips      []xds.VIPDomains
		thenOutbounds []*xds.Outbound
	}
	DescribeTable("compute outbounds",
		func(tc outboundTestCase) {
			vobView, err := vips.NewVirtualOutboundView(tc.whenOutbounds)
			Expect(err).ToNot(HaveOccurred())

			vips, outbounds := topology.VIPOutbounds(vobView, "mesh", 80)

			Expect(vips).To(Equal(tc.thenVips))
			Expect(outbounds).To(Equal(tc.thenOutbounds))
		},
		Entry("empty", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{},
		}),
		Entry("host port backcompat", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewHostEntry("example.com"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{Port: 1234, TagSet: map[string]string{mesh_proto.ServiceTag: "foo"}},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.1", Domains: []string{"example.com"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 1234, Tags: map[string]string{mesh_proto.ServiceTag: "foo"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "foo"}}},
			},
		}),
		Entry("host port no backcompat when using port 80", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewHostEntry("example.com"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{Port: 81, TagSet: map[string]string{mesh_proto.ServiceTag: "bar"}},
						{Port: 80, TagSet: map[string]string{mesh_proto.ServiceTag: "foo"}},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.1", Domains: []string{"example.com"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 81, Tags: map[string]string{mesh_proto.ServiceTag: "bar"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "foo"}}},
			},
		}),
		Entry("with no address doesn't generate vips or outbounds", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("example"): {
					Address: "",
					Outbounds: []vips.OutboundEntry{
						{Port: 1234, TagSet: map[string]string{mesh_proto.ServiceTag: "foo"}},
					},
				},
			},
		}),
		Entry("service generates hostnames", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("example"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{TagSet: map[string]string{mesh_proto.ServiceTag: "example"}},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.1", Domains: []string{"example.mesh"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "example"}}},
			},
		}),
		Entry("service with port add backcompat", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("example"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{TagSet: map[string]string{mesh_proto.ServiceTag: "example"}, Port: 1234},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.1", Domains: []string{"example.mesh"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 1234, Tags: map[string]string{mesh_proto.ServiceTag: "example"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "example"}}},
			},
		}),
		Entry("service normalizes hostnames", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewServiceEntry("example_svc_80"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{TagSet: map[string]string{mesh_proto.ServiceTag: "example_svc_80"}},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.1", Domains: []string{"example_svc_80.mesh", "example.svc.80.mesh"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 80, Tags: map[string]string{mesh_proto.ServiceTag: "example_svc_80"}}},
			},
		}),
		Entry("multi outbounds work", outboundTestCase{
			whenOutbounds: map[vips.HostnameEntry]vips.VirtualOutbound{
				vips.NewFqdnEntry("my-foo-service-generated.mesh"): {
					Address: "240.0.0.1",
					Outbounds: []vips.OutboundEntry{
						{Port: 1234, TagSet: map[string]string{mesh_proto.ServiceTag: "foo", "version": "1"}},
						{Port: 1235, TagSet: map[string]string{mesh_proto.ServiceTag: "foo", "version": "2"}},
					},
				},
				vips.NewFqdnEntry("my-bar-service-generated.mesh"): {
					Address: "240.0.0.2",
					Outbounds: []vips.OutboundEntry{
						{Port: 1234, TagSet: map[string]string{mesh_proto.ServiceTag: "bar", "version": "1"}},
						{Port: 1235, TagSet: map[string]string{mesh_proto.ServiceTag: "bar", "version": "2"}},
					},
				},
			},
			thenVips: []xds.VIPDomains{
				{Address: "240.0.0.2", Domains: []string{"my-bar-service-generated.mesh"}},
				{Address: "240.0.0.1", Domains: []string{"my-foo-service-generated.mesh"}},
			},
			thenOutbounds: []*xds.Outbound{
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.2", Port: 1234, Tags: map[string]string{mesh_proto.ServiceTag: "bar", "version": "1"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.2", Port: 1235, Tags: map[string]string{mesh_proto.ServiceTag: "bar", "version": "2"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 1234, Tags: map[string]string{mesh_proto.ServiceTag: "foo", "version": "1"}}},
				{LegacyOutbound: &mesh_proto.Dataplane_Networking_Outbound{Address: "240.0.0.1", Port: 1235, Tags: map[string]string{mesh_proto.ServiceTag: "foo", "version": "2"}}},
			},
		}),
	)
})
