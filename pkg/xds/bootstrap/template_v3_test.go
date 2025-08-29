package bootstrap

import (
	"errors"
	"net"

	envoy_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_grpc_credentials_v3 "github.com/envoyproxy/go-control-plane/envoy/config/grpc_credential/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	xds_config "github.com/kumahq/kuma/pkg/config/xds"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("dnsLookupFamilyFromXdsHost", func() {
	It("should return AUTO when IPv6 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv6loopback}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO for localhost", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("localhost", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO when both IPv6 and IPv4 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv6loopback, net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return IPV4 when only IPv4 found", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{net.IPv4(127, 0, 0, 1)}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_V4_ONLY))
	})

	It("should return AUTO (default) when no ips returned", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return []net.IP{}, nil
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})

	It("should return AUTO when error occurs", func() {
		// given
		lookupFn := func(host string) ([]net.IP, error) {
			return nil, errors.New("could not resolve hostname")
		}

		// when
		result := dnsLookupFamilyFromXdsHost("example.com", lookupFn)

		// then
		Expect(result).To(Equal(envoy_cluster_v3.Cluster_AUTO))
	})
})

var _ = Describe("genConfig", func() {
	It("should has google grpc and no initial metadata when use path enabled and path provided", func() {
		// given
		params := configParameters{
			Id:                 "default.backend",
			Version:            &mesh_proto.Version{},
			XdsHost:            "control-plane.host",
			XdsPort:            5678,
			DataplaneToken:     "token",
			DataplaneTokenPath: "/path/to/file",
			CertBytes:          []byte{0x00},
			HdsEnabled:         true,
		}

		// when
		result, err := genConfig(params, xds_config.Proxy{}, true, nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.GetDynamicResources().GetAdsConfig().GetGrpcServices()).To(HaveLen(1))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].GetGoogleGrpc().CallCredentials[0]).To(Equal(
			&envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials{
				CredentialSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
					FromPlugin: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
						Name: "envoy.grpc_credentials.file_based_metadata",
						ConfigType: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
							TypedConfig: util_proto.MustMarshalAny(&envoy_grpc_credentials_v3.FileBasedMetadataConfig{
								SecretData: &envoy_core_v3.DataSource{
									Specifier: &envoy_core_v3.DataSource_Filename{Filename: "/path/to/file"},
								},
							}),
						},
					},
				},
			},
		))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].GetInitialMetadata()).To(BeEmpty())
		Expect(result.GetHdsConfig().GrpcServices[0].GetGoogleGrpc().CallCredentials[0]).To(Equal(
			&envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials{
				CredentialSpecifier: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_FromPlugin{
					FromPlugin: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin{
						Name: "envoy.grpc_credentials.file_based_metadata",
						ConfigType: &envoy_core_v3.GrpcService_GoogleGrpc_CallCredentials_MetadataCredentialsFromPlugin_TypedConfig{
							TypedConfig: util_proto.MustMarshalAny(&envoy_grpc_credentials_v3.FileBasedMetadataConfig{
								SecretData: &envoy_core_v3.DataSource{
									Specifier: &envoy_core_v3.DataSource_Filename{Filename: "/path/to/file"},
								},
							}),
						},
					},
				},
			},
		))
		Expect(result.HdsConfig.GrpcServices[0].GetInitialMetadata()).To(BeEmpty())
	})

	It("should has initial metadata when usePath disabled", func() {
		// given
		params := configParameters{
			Id:                 "default.backend",
			Version:            &mesh_proto.Version{},
			XdsHost:            "control-plane.host",
			XdsPort:            5678,
			DataplaneToken:     "token",
			DataplaneTokenPath: "/path/to/file",
			CertBytes:          []byte{0x00},
			HdsEnabled:         true,
		}

		// when
		result, err := genConfig(params, xds_config.Proxy{}, false, nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.GetDynamicResources().GetAdsConfig().GetGrpcServices()).To(HaveLen(1))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].GetEnvoyGrpc()).To(Equal(
			&envoy_core_v3.GrpcService_EnvoyGrpc{
				ClusterName: "ads_cluster",
			},
		))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].InitialMetadata[0]).To(Equal(
			&envoy_core_v3.HeaderValue{Key: "authorization", Value: "token"},
		))
		Expect(result.GetHdsConfig().GrpcServices[0].GetEnvoyGrpc()).To(Equal(
			&envoy_core_v3.GrpcService_EnvoyGrpc{
				ClusterName: "ads_cluster",
			},
		))
		Expect(result.HdsConfig.GrpcServices[0].InitialMetadata[0]).To(Equal(
			&envoy_core_v3.HeaderValue{Key: "authorization", Value: "token"},
		))
	})

	It("should use envoy grpc and has initial metadata when usePath enabled but no path", func() {
		// given
		params := configParameters{
			Id:                 "default.backend",
			Version:            &mesh_proto.Version{},
			XdsHost:            "control-plane.host",
			XdsPort:            5678,
			DataplaneToken:     "token",
			DataplaneTokenPath: "",
			CertBytes:          []byte{0x00},
			HdsEnabled:         true,
		}

		// when
		result, err := genConfig(params, xds_config.Proxy{}, true, nil)

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(result.GetDynamicResources().GetAdsConfig().GetGrpcServices()).To(HaveLen(1))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].GetEnvoyGrpc()).To(Equal(
			&envoy_core_v3.GrpcService_EnvoyGrpc{
				ClusterName: "ads_cluster",
			},
		))
		Expect(result.GetDynamicResources().GetAdsConfig().GrpcServices[0].InitialMetadata[0]).To(Equal(
			&envoy_core_v3.HeaderValue{Key: "authorization", Value: "token"},
		))
		Expect(result.GetHdsConfig().GrpcServices[0].GetEnvoyGrpc()).To(Equal(
			&envoy_core_v3.GrpcService_EnvoyGrpc{
				ClusterName: "ads_cluster",
			},
		))
		Expect(result.HdsConfig.GrpcServices[0].InitialMetadata[0]).To(Equal(
			&envoy_core_v3.HeaderValue{Key: "authorization", Value: "token"},
		))
	})
})
