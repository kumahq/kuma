package meshroute

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	core_xds "github.com/kumahq/kuma/v3/pkg/core/xds"
)

var _ = Describe("classifyMZMSEndpointZones", func() {
	endpoint := func(zone string) core_xds.Endpoint {
		if zone == "" {
			return core_xds.Endpoint{}
		}
		return core_xds.Endpoint{Locality: &core_xds.Locality{Zone: zone}}
	}

	type testCase struct {
		endpoints           []core_xds.Endpoint
		zonesWithProxy      map[string]bool
		localZone           string
		expectedLegacyZones []string
		expectedHasDefault  bool
	}

	DescribeTable("partitions endpoints by the SNI format their zone expects",
		func(given testCase) {
			legacyZones, hasDefaultSNIEndpoint := classifyMZMSEndpointZones(given.endpoints, given.zonesWithProxy, given.localZone)

			Expect(legacyZones).To(Equal(given.expectedLegacyZones))
			Expect(hasDefaultSNIEndpoint).To(Equal(given.expectedHasDefault))
		},
		Entry("local endpoints never count as legacy, even without a local MeshZoneAddress", testCase{
			endpoints:           []core_xds.Endpoint{endpoint("west"), endpoint("")},
			zonesWithProxy:      map[string]bool{},
			localZone:           "west",
			expectedLegacyZones: nil,
			expectedHasDefault:  true,
		}),
		Entry("remote zones with a MeshZoneAddress use the default SNI", testCase{
			endpoints:           []core_xds.Endpoint{endpoint("west"), endpoint("east")},
			zonesWithProxy:      map[string]bool{"east": true},
			localZone:           "west",
			expectedLegacyZones: nil,
			expectedHasDefault:  true,
		}),
		Entry("remote zones without a MeshZoneAddress are legacy", testCase{
			endpoints:           []core_xds.Endpoint{endpoint("east"), endpoint("north")},
			zonesWithProxy:      map[string]bool{"north": true},
			localZone:           "west",
			expectedLegacyZones: []string{"east"},
			expectedHasDefault:  true,
		}),
		Entry("legacy zones are sorted and deduplicated", testCase{
			endpoints:           []core_xds.Endpoint{endpoint("south"), endpoint("east"), endpoint("south")},
			zonesWithProxy:      map[string]bool{},
			localZone:           "west",
			expectedLegacyZones: []string{"east", "south"},
			expectedHasDefault:  false,
		}),
	)
})
