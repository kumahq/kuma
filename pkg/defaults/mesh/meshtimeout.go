package mesh

import (
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/core/resources/model"
	policies_defaults "github.com/kumahq/kuma/v3/pkg/plugins/policies/core/defaults"
	"github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtimeout/api/v1alpha1"
	envoy_common "github.com/kumahq/kuma/v3/pkg/xds/envoy"
)

// defaultMeshTimeoutResource and defaultMeshTimeoutToResource are kept as
// separate resources: 'rules' (inbound) and 'to' (outbound) are mutually
// exclusive on a single MeshTimeout (see validator.go), so the mesh-wide
// inbound and outbound defaults can't be expressed on one resource.
var defaultMeshTimeoutResource = func() model.Resource {
	const factor = 2
	return &v1alpha1.MeshTimeoutResource{
		Spec: &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
			},

			// bigger than outbound side timeouts or disabled.
			Rules: &[]v1alpha1.Rule{
				{
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
		},
	}
}

var defaultMeshTimeoutToResource = func() model.Resource {
	return &v1alpha1.MeshTimeoutResource{
		Spec: &v1alpha1.MeshTimeout{
			TargetRef: &common_api.TargetRef{
				Kind: common_api.Mesh,
			},
			To: &[]v1alpha1.To{
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

// DefaultInboundTimeout returns timeouts for the inbound side. This resource is not created
// in the store. It's used directly in InboundProxyGenerator. In the future, it could be replaced
// with a new InboundTimeout policy. The main idea around these values is to have them either
// bigger than outbound side timeouts or disabled.
var DefaultInboundTimeout = func() envoy_common.Timeouts {
	const factor = 2

	return envoy_common.Timeouts{
		Connect:        factor * policies_defaults.DefaultConnectTimeout,
		TcpIdle:        factor * policies_defaults.DefaultIdleTimeout,
		HttpIdle:       factor * policies_defaults.DefaultIdleTimeout,
		HttpStreamIdle: factor * policies_defaults.DefaultStreamIdleTimeout,
	}
}
