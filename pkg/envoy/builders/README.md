# Envoy Builders

## Motivation

1. Reusable code across policy plugins
2. Compact notation improves readability. The logic is not spread across multiple functions or polluted with if/else branches.
See the example from MeshLoadBalancingStrategy `plugin.go`:

```go
func listenerConfigurer(rctx *rules_outbound.ResourceContext[api.Conf]) Configurer[envoy_listener.Listener] {
	return bldrs_listener.FilterChains(filterChainConfigurer(rctx))
}

func clusterConfigurer(conf api.Conf) Configurer[envoy_cluster.Cluster] {
	return func(cluster *envoy_cluster.Cluster) error {
		return NewModifier(cluster).
			Configure(If(shouldUseLocalityWeightedLb(conf), bldrs_clusters.LocalityWeightedLbConfigurer())).
			Configure(IfNotNil(conf.LoadBalancer, loadBalancerConfigurer)).
Modify()
	}
}

func claConfigurer(conf api.Conf, tags mesh_proto.MultiValueTagSet, localZone string, egressEnabled bool, origin string) Configurer[envoy_endpoint.ClusterLoadAssignment] {
	return func(cla *envoy_endpoint.ClusterLoadAssignment) error {
		atLeastOneLocalityGroup := conf.LocalityAwareness != nil && (conf.LocalityAwareness.LocalZone != nil || conf.LocalityAwareness.CrossZone != nil)
		isLocalityAware := conf.LocalityAwareness == nil || !pointer.Deref(conf.LocalityAwareness.Disabled)
		return NewModifier(cla).
			Configure(bldrs_endpoint.NonLocalPriority(isLocalityAware, localZone)).
			Configure(If(atLeastOneLocalityGroup, bldrs_endpoint.Endpoints(NewEndpoints(cla.Endpoints, tags, pointer.To(conf), localZone, egressEnabled, origin)))).
			Configure(If(atLeastOneLocalityGroup, bldrs_endpoint.OverprovisioningFactor(overprovisioningFactor(conf)))).
			Modify()
	}
}
```

## How to contribute builders?

1. When adding new package to `builders` follow the structure from [envoy/api](https://github.com/envoyproxy/envoy/tree/main/api/envoy/config)

2. Never depend on any code from `pkg/plugins/policies`

### Example

Let's say you want to build envoy proto `MetadataMatcher`:

```go
type MetadataMatcher struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The filter name to retrieve the Struct from the Metadata.
	Filter string `protobuf:"bytes,1,opt,name=filter,proto3" json:"filter,omitempty"`
	// The path to retrieve the Value from the Struct.
	Path []*MetadataMatcher_PathSegment `protobuf:"bytes,2,rep,name=path,proto3" json:"path,omitempty"`
	// The MetadataMatcher is matched if the value retrieved by path is matched to this value.
	Value *ValueMatcher `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	// If true, the match result will be inverted.
	Invert bool `protobuf:"varint,4,opt,name=invert,proto3" json:"invert,omitempty"`
}
```

Start with creating `NewXBuilder()` method, so that client code doesn't have to deal with generics:

```go
func NewMetadataBuilder() *Builder[matcherv3.MetadataMatcher] {
	return &Builder[matcherv3.MetadataMatcher]{}
}
```

Next, create functions that return `Configurer`. 
`Configurer` is just a function that sets some fields on `MetadataMatcher`, i.e:

```go
func Key(key string) Configurer[matcherv3.MetadataMatcher] {
	return func(m *matcherv3.MetadataMatcher) error {
		m.Path = []*matcherv3.MetadataMatcher_PathSegment{
			{
				Segment: &matcherv3.MetadataMatcher_PathSegment_Key{
					Key: key,
				},
			},
		}
		return nil
	}
}

func ExactValue(value string) Configurer[matcherv3.MetadataMatcher] {
	return func(m *matcherv3.MetadataMatcher) error {
		m.Value = &matcherv3.ValueMatcher{
			MatchPattern: &matcherv3.ValueMatcher_StringMatch{
				StringMatch: &matcherv3.StringMatcher{
					MatchPattern: &matcherv3.StringMatcher_Exact{
						Exact: value,
					},
				},
			},
		}
		return nil
	}
}
```

For `Key` and `ExactValue` functions try to keep the list of args consisting of basic Go types.

Now, the client code can use integrate it in building of some other object that has `MetadataMatcher`:

```go
SomeBuilderWithMetadataMatcher(...).
  Configure(bldrs_matcher.NewMetadataBuilder().
    Configure(bldrs_matcher.Key(myKey)).
    Configure(bldrs_matcher.ExactValue(myValue)))
```
