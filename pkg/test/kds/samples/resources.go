package samples

import (
	common_api "github.com/kumahq/kuma/v3/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/v3/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/v3/api/system/v1alpha1"
	meshaccesslog "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtrafficpermissions "github.com/kumahq/kuma/v3/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/v3/pkg/util/proto"
)

var (
	Mesh1 = &mesh_proto.Mesh{
		Mtls: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-1",
			Backends: []*mesh_proto.CertificateAuthorityBackend{
				{
					Name: "ca-1",
					Type: "builtin",
				},
			},
		},
	}
	Mesh2 = &mesh_proto.Mesh{
		Mtls: &mesh_proto.Mesh_Mtls{
			EnabledBackend: "ca-2",
			Backends: []*mesh_proto.CertificateAuthorityBackend{
				{
					Name: "ca-2",
					Type: "builtin",
				},
			},
		},
	}
	Dataplane = &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.1",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
				Port: 1212,
			}},
			Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 1213,
					BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
						Kind: "MeshService",
						Name: "web",
						Port: 1213,
					},
				},
			},
		},
	}
	GatewayDataplane = &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Gateway: &mesh_proto.Dataplane_Networking_Gateway{
				Tags: map[string]string{
					mesh_proto.ServiceTag: "gateway",
				},
				Type: mesh_proto.Dataplane_Networking_Gateway_DELEGATED,
			},
			Address: "192.168.0.1",
			Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 1213,
					BackendRef: &mesh_proto.Dataplane_Networking_Outbound_BackendRef{
						Kind: "MeshService",
						Name: "web",
						Port: 1213,
					},
				},
			},
		},
	}
	DataplaneInsight = &mesh_proto.DataplaneInsight{
		MTLS: &mesh_proto.DataplaneInsight_MTLS{
			CertificateRegenerations: 3,
		},
	}
	ServiceInsight = &mesh_proto.ServiceInsight{
		Services: map[string]*mesh_proto.ServiceInsight_Service{},
	}
	Secret2 = &system_proto.Secret{
		Data: util_proto.Bytes([]byte("secret")),
	}
	Secret = &system_proto.Secret{
		Data: util_proto.Bytes([]byte("secret key")),
	}
	GlobalSecret = &system_proto.Secret{
		Data: util_proto.Bytes([]byte("global secret key")),
	}
	Config = &system_proto.Config{
		Config: "sample config",
	}
	MeshTrafficPermission = &meshtrafficpermissions.MeshTrafficPermission{
		TargetRef: &common_api.TargetRef{
			Kind: "Mesh",
		},
		Rules: &[]meshtrafficpermissions.Rule{
			{
				Default: meshtrafficpermissions.RuleConf{
					Allow: &[]common_api.Match{
						{
							SpiffeID: &common_api.SpiffeIDMatch{
								Type:  common_api.PrefixMatchType,
								Value: "spiffe://default",
							},
						},
					},
				},
			},
		},
	}
	MeshAccessLog = &meshaccesslog.MeshAccessLog{
		TargetRef: &common_api.TargetRef{
			Kind: "Mesh",
		},
		Rules: &[]meshaccesslog.Rule{
			{
				Default: meshaccesslog.Conf{
					Backends: &[]meshaccesslog.Backend{
						{
							Type: meshaccesslog.FileBackendType,
							File: &meshaccesslog.FileBackend{
								Path: "/dev/null",
							},
						},
					},
				},
			},
		},
	}
)
