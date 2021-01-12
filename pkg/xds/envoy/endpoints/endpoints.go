package endpoints

import (
	"errors"

	"github.com/golang/protobuf/proto"

	core_xds "github.com/kumahq/kuma/pkg/core/xds"
	envoy_common "github.com/kumahq/kuma/pkg/xds/envoy"
	endpoints_v2 "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v2"
	endpoints_v3 "github.com/kumahq/kuma/pkg/xds/envoy/endpoints/v3"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32, apiVersion envoy_common.APIVersion) (proto.Message, error) {
	switch apiVersion {
	case envoy_common.APIV2:
		return endpoints_v2.CreateStaticEndpoint(clusterName, address, port), nil
	case envoy_common.APIV3:
		return endpoints_v3.CreateStaticEndpoint(clusterName, address, port), nil
	default:
		return nil, errors.New("unknown API")
	}
}

func CreateClusterLoadAssignment(clusterName string, endpoints []core_xds.Endpoint, apiVersion envoy_common.APIVersion) (proto.Message, error) {
	switch apiVersion {
	case envoy_common.APIV2:
		return endpoints_v2.CreateClusterLoadAssignment(clusterName, endpoints), nil
	case envoy_common.APIV3:
		return endpoints_v3.CreateClusterLoadAssignment(clusterName, endpoints), nil
	default:
		return nil, errors.New("unknown API")
	}
}
