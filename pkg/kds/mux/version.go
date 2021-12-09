package mux

import (
	"context"

	envoy_api_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core_v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_core_v3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_sd_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/pkg/errors"
	"google.golang.org/grpc/metadata"
)

const (
	KDSVersionHeaderKey = "kds-version"
	KDSVersionV2        = "v2"
	KDSVersionV3        = "v3"
)

func KDSVersion(ctx context.Context) string {
	md, found := metadata.FromIncomingContext(ctx)
	if found {
		version := md.Get(KDSVersionHeaderKey)
		if len(version) == 1 {
			return version[0]
		}
	}

	md, found = metadata.FromOutgoingContext(ctx)
	if found {
		version := md.Get(KDSVersionHeaderKey)
		if len(version) == 1 {
			return version[0]
		}
	}

	return KDSVersionV2
}

func UnsupportedKDSVersion(version string) error {
	return errors.Errorf("invalid KDS version %s. Supported versions %s %s", version, KDSVersionV2, KDSVersionV3)
}

func DiscoveryRequestV2(request *envoy_sd_v3.DiscoveryRequest) *envoy_api_v2.DiscoveryRequest {
	return &envoy_api_v2.DiscoveryRequest{
		VersionInfo: request.VersionInfo,
		Node: &envoy_core_v2.Node{
			Id:       request.Node.Id,
			Metadata: request.Node.Metadata,
		},
		ResourceNames: request.ResourceNames,
		TypeUrl:       request.TypeUrl,
		ResponseNonce: request.ResponseNonce,
		ErrorDetail:   request.ErrorDetail,
	}
}

func DiscoveryRequestV3(request *envoy_api_v2.DiscoveryRequest) *envoy_sd_v3.DiscoveryRequest {
	return &envoy_sd_v3.DiscoveryRequest{
		VersionInfo: request.VersionInfo,
		Node: &envoy_core_v3.Node{
			Id:       request.Node.Id,
			Metadata: request.Node.Metadata,
		},
		ResourceNames: request.ResourceNames,
		TypeUrl:       request.TypeUrl,
		ResponseNonce: request.ResponseNonce,
		ErrorDetail:   request.ErrorDetail,
	}
}

func DiscoveryResponseV2(response *envoy_sd_v3.DiscoveryResponse) *envoy_api_v2.DiscoveryResponse {
	return &envoy_api_v2.DiscoveryResponse{
		VersionInfo: response.VersionInfo,
		Resources:   response.Resources,
		TypeUrl:     response.TypeUrl,
		Nonce:       response.Nonce,
		ControlPlane: &envoy_core_v2.ControlPlane{
			Identifier: response.ControlPlane.Identifier,
		},
	}
}

func DiscoveryResponseV3(response *envoy_api_v2.DiscoveryResponse) *envoy_sd_v3.DiscoveryResponse {
	return &envoy_sd_v3.DiscoveryResponse{
		VersionInfo: response.VersionInfo,
		Resources:   response.Resources,
		TypeUrl:     response.TypeUrl,
		Nonce:       response.Nonce,
		ControlPlane: &envoy_core_v3.ControlPlane{
			Identifier: response.ControlPlane.Identifier,
		},
	}
}
