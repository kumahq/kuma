package clusters

import (
	envoy_cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	envoy_upstream_http "github.com/envoyproxy/go-control-plane/envoy/extensions/upstreams/http/v3"
	"google.golang.org/protobuf/types/known/anypb"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

func UpdateCommonHttpProtocolOptions(cluster *envoy_cluster.Cluster, fn func(*envoy_upstream_http.HttpProtocolOptions)) error {
	if cluster.TypedExtensionProtocolOptions == nil {
		cluster.TypedExtensionProtocolOptions = map[string]*anypb.Any{}
	}
	options := &envoy_upstream_http.HttpProtocolOptions{}
	if a := cluster.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"]; a != nil {
		if err := util_proto.UnmarshalAnyTo(a, options); err != nil {
			return err
		}
	}

	fn(options)

	pbst, err := util_proto.MarshalAnyDeterministic(options)
	if err != nil {
		return err
	}
	cluster.TypedExtensionProtocolOptions["envoy.extensions.upstreams.http.v3.HttpProtocolOptions"] = pbst
	return nil
}
