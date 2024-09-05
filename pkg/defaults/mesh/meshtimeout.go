package mesh

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	policies_defaults "github.com/kumahq/kuma/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var defaultMeshTimeoutResource = func() model.Resource {
	const factor = 2
	return &v1alpha1.MeshTimeoutResource{
		Spec: &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
				ProxyTypes: []common_api.TargetRefProxyType{
					common_api.Sidecar,
				},
			},

			// bigger than outbound side timeouts or disabled.
			From: []v1alpha1.From{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						ConnectionTimeout: &kube_meta.Duration{
							Duration: factor * policies_defaults.DefaultConnectTimeout,
						},
						IdleTimeout: &kube_meta.Duration{
							Duration: factor * policies_defaults.DefaultIdleTimeout,
						},
						Http: &v1alpha1.Http{
							RequestTimeout: &kube_meta.Duration{
								Duration: 0,
							},
							StreamIdleTimeout: &kube_meta.Duration{
								Duration: factor * policies_defaults.DefaultStreamIdleTimeout,
							},
							MaxStreamDuration: &kube_meta.Duration{
								Duration: 0,
							},
						},
					},
				},
			},
			To: []v1alpha1.To{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						ConnectionTimeout: &kube_meta.Duration{
							Duration: policies_defaults.DefaultConnectTimeout,
						},
						IdleTimeout: &kube_meta.Duration{
							Duration: policies_defaults.DefaultIdleTimeout,
						},
						Http: &v1alpha1.Http{
							RequestTimeout: &kube_meta.Duration{
								Duration: policies_defaults.DefaultRequestTimeout,
							},
							StreamIdleTimeout: &kube_meta.Duration{
								Duration: policies_defaults.DefaultStreamIdleTimeout,
							},
						},
					},
				},
			},
		},
	}
}

var defaulMeshGatewaysTimeoutResource = func() model.Resource {
	return &v1alpha1.MeshTimeoutResource{
		Spec: &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
				ProxyTypes: []common_api.TargetRefProxyType{
					common_api.Gateway,
				},
			},
			From: []v1alpha1.From{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						IdleTimeout: &kube_meta.Duration{
							Duration: policies_defaults.DefaultGatewayIdleTimeout,
						},
						Http: &v1alpha1.Http{
							StreamIdleTimeout: &kube_meta.Duration{
								Duration: policies_defaults.DefaultGatewayStreamIdleTimeout,
							},
							RequestHeadersTimeout: &kube_meta.Duration{
								Duration: policies_defaults.DefaultGatewayRequestHeadersTimeout,
							},
						},
					},
				},
			},
			To: []v1alpha1.To{
				{
					TargetRef: common_api.TargetRef{
						Kind: common_api.Mesh,
					},
					Default: v1alpha1.Conf{
						IdleTimeout: &kube_meta.Duration{
							Duration: policies_defaults.DefaultIdleTimeout,
						},
						Http: &v1alpha1.Http{
							StreamIdleTimeout: &kube_meta.Duration{
								Duration: policies_defaults.DefaultGatewayStreamIdleTimeout,
							},
						},
					},
				},
			},
		},
	}
}

// DefaultInboundTimeout returns timeouts for the inbound side. This resource is not created
// in the store. It's used directly in InboundProxyGenerator. In the future, it could be replaced
// with a new InboundTimeout policy. The main idea around these values is to have them either
// bigger than outbound side timeouts or disabled.
var DefaultInboundTimeout = func() *mesh_proto.Timeout_Conf {
	const factor = 2

	return &mesh_proto.Timeout_Conf{
		ConnectTimeout: util_proto.Duration(factor * policies_defaults.DefaultConnectTimeout),
		Tcp: &mesh_proto.Timeout_Conf_Tcp{
			IdleTimeout: util_proto.Duration(factor * policies_defaults.DefaultIdleTimeout),
		},
		Http: &mesh_proto.Timeout_Conf_Http{
			RequestTimeout:    util_proto.Duration(0),
			IdleTimeout:       util_proto.Duration(factor * policies_defaults.DefaultIdleTimeout),
			StreamIdleTimeout: util_proto.Duration(factor * policies_defaults.DefaultStreamIdleTimeout),
			MaxStreamDuration: util_proto.Duration(0),
		},
	}
}
