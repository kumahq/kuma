package bootstrap

import (
	"errors"
	"net"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dnsLookupFamilyFromXdsHost", func() {
	It("should return AUTO when IPv6 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv6loopback}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO for localhost", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("localhost", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO when both IPv6 and IPv4 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv6loopback, net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return IPV4 when only IPv4 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_V4_ONLY))
	})

	It("should return AUTO (default) when no ips returned", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO when error occurs", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return nil, errors.New("could not resolve hostname")
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		//
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})
})
