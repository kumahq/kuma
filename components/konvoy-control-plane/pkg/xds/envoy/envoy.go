package envoy

import (
	"time"

	"github.com/gogo/protobuf/types"

	"fmt"

	mesh_core "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/core/resources/apis/mesh"
	util_error "github.com/Kong/konvoy/components/konvoy-control-plane/pkg/util/error"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/endpoint"
	"github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	rbac "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/rbac/v2"
	tcp "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/tcp_proxy/v2"
	rbac_config "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v2"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	"github.com/envoyproxy/go-control-plane/pkg/util"
)

func CreateStaticEndpoint(clusterName string, address string, port uint32) *v2.ClusterLoadAssignment {
	return &v2.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []endpoint.LocalityLbEndpoints{{
			LbEndpoints: []endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: &core.Address{
							Address: &core.Address_SocketAddress{
								SocketAddress: &core.SocketAddress{
									Protocol: core.TCP,
									Address:  address,
									PortSpecifier: &core.SocketAddress_PortValue{
										PortValue: port,
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func CreateLocalCluster(clusterName string, address string, port uint32) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       5 * time.Second,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_STATIC},
		LoadAssignment:       CreateStaticEndpoint(clusterName, address, port),
	}
}

func CreatePassThroughCluster(clusterName string) *v2.Cluster {
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       5 * time.Second,
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_ORIGINAL_DST},
		LbPolicy:             v2.Cluster_ORIGINAL_DST_LB,
	}
}

// get listenerName and permissions list and returns a constructed RBAC rule
func createRbacRule(listenerName string, permissions *mesh_core.TrafficPermissionResourceList) *rbac.RBAC {
	policies := make(map[string]*rbac_config.Policy, len(permissions.Items))

	/*
				- For each traffic permission,
					- create a policy, whose name is the traffic permission name
						- for each policy
							- create a Permissions object, with a single permission inside
								- Any: true
							- create a Principals object, with a single principal inside


				- name: envoy.filters.network.rbac
		          rules:
		            action: ALLOW
		            policies:
		              "traffic-permission-name":
		                permissions:
		                - any: true
		                principals:
		                - authenticated:
		                    principal_name:
		                      exact: "spiffe://<mesh name from traffic permission>/<service tag>"
	*/

	principals := []*rbac_config.Principal{}

	for _, trafficPermission := range permissions.Items {

		meta := trafficPermission.GetMeta()
		policyName := fmt.Sprintf("%s.%s", meta.GetNamespace(), meta.GetName())

		// build principals list: one per sources/destinations rule
		for _, rule := range trafficPermission.Spec.Rules {
			principal := &rbac_config.Principal{
				Identifier: &rbac_config.Principal_Authenticated_{
					Authenticated: &rbac_config.Principal_Authenticated{
						PrincipalName: &matcher.StringMatcher{
							MatchPattern: &matcher.StringMatcher_Exact{
								Exact: "spiffe://" + meta.GetMesh() + "/" + rule.Sources[0].Match["service"],
							},
						},
					},
				},
			}

			principals = append(principals, principal)
		}

		// construct policies: one per traffic permission (name is namespace.name)
		policies[policyName] = &rbac_config.Policy{
			Permissions: []*rbac_config.Permission{
				&rbac_config.Permission{
					Rule: &rbac_config.Permission_Any{
						Any: true,
					},
				},
			},
			Principals: principals,
		}
	}

	rbacRule := &rbac.RBAC{
		Rules: &rbac_config.RBAC{
			Action:   rbac_config.RBAC_ALLOW,
			Policies: policies,
		},
		StatPrefix: listenerName,
	}

	return rbacRule
}

func CreateInboundListener(listenerName string, address string, port uint32, clusterName string, virtual bool, permissions *mesh_core.TrafficPermissionResourceList) *v2.Listener {
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}

	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)

	rbacRule := createRbacRule(listenerName, permissions)
	rbacMarshalled, err := types.MarshalAny(rbacRule)
	util_error.MustNot(err)

	listener := &v2.Listener{
		Name: listenerName,
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []listener.FilterChain{{
			Filters: []listener.Filter{
				{
					Name: "envoy.filters.network.rbac", // TODO find out where const is in go-control-plane
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: rbacMarshalled,
					},
				},
				{
					Name: util.TCPProxy,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				},
			},
		}},
	}
	if virtual {
		// TODO(yskopets): What is the up-to-date alternative ?
		listener.DeprecatedV1 = &v2.Listener_DeprecatedV1{
			BindToPort: &types.BoolValue{Value: false},
		}
	}
	return listener
}

func CreateCatchAllListener(listenerName string, address string, port uint32, clusterName string) *v2.Listener {
	config := &tcp.TcpProxy{
		StatPrefix: clusterName,
		ClusterSpecifier: &tcp.TcpProxy_Cluster{
			Cluster: clusterName,
		},
	}

	pbst, err := types.MarshalAny(config)
	util_error.MustNot(err)

	return &v2.Listener{
		Name: listenerName,
		Address: core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.TCP,
					Address:  address,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []listener.FilterChain{{
			Filters: []listener.Filter{
				{
					Name: util.TCPProxy,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				},
			},
		}},
		// TODO(yskopets): What is the up-to-date alternative ?
		UseOriginalDst: &types.BoolValue{Value: true},
		// TODO(yskopets): Apparently, `envoy.listener.original_dst` has different effect than `UseOriginalDst`
		//ListenerFilters: []listener.ListenerFilter{{
		//	Name: util.OriginalDestination,
		//}},
	}
}
