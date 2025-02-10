package samples

import (
	"time"

	common_api "github.com/kumahq/kuma/api/common/v1alpha1"
	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	meshaccesslog "github.com/kumahq/kuma/pkg/plugins/policies/meshaccesslog/api/v1alpha1"
	meshtrafficpermissions "github.com/kumahq/kuma/pkg/plugins/policies/meshtrafficpermission/api/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
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
	FaultInjection = &mesh_proto.FaultInjection{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
				"tag0":                "version0",
				"tag1":                "version1",
				"tag2":                "version2",
				"tag3":                "version3",
				"tag4":                "version4",
				"tag5":                "version5",
				"tag6":                "version6",
				"tag7":                "version7",
				"tag8":                "version8",
				"tag9":                "version9",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag:  "*",
				mesh_proto.ProtocolTag: "http",
			},
		}},
		Conf: &mesh_proto.FaultInjection_Conf{
			Abort: &mesh_proto.FaultInjection_Conf_Abort{
				Percentage: util_proto.Double(90),
				HttpStatus: util_proto.UInt32(404),
			},
		},
	}
	Dataplane = &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Address: "192.168.0.1",
			Inbound: []*mesh_proto.Dataplane_Networking_Inbound{{
				Port: 1212,
				Tags: map[string]string{
					mesh_proto.ZoneTag:    "kuma-1",
					mesh_proto.ServiceTag: "backend",
				},
			}},
			Outbound: []*mesh_proto.Dataplane_Networking_Outbound{
				{
					Port: 1213,
					Tags: map[string]string{
						mesh_proto.ServiceTag:  "web",
						mesh_proto.ProtocolTag: "http",
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
					Tags: map[string]string{
						mesh_proto.ServiceTag:  "web",
						mesh_proto.ProtocolTag: "http",
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
	ZoneIngress = &mesh_proto.ZoneIngress{
		Networking: &mesh_proto.ZoneIngress_Networking{
			Address:           "127.0.0.1",
			Port:              80,
			AdvertisedAddress: "192.168.0.1",
			AdvertisedPort:    10001,
		},
		AvailableServices: []*mesh_proto.ZoneIngress_AvailableService{
			{
				Tags: map[string]string{
					mesh_proto.ServiceTag: "backend",
				},
			},
		},
	}
	ZoneIngressInsight = &mesh_proto.ZoneIngressInsight{
		Subscriptions: []*mesh_proto.DiscoverySubscription{{
			Id: "1",
		}},
	}
	ZoneEgress = &mesh_proto.ZoneEgress{
		Networking: &mesh_proto.ZoneEgress_Networking{
			Address: "127.0.0.1",
			Port:    80,
		},
	}
	ZoneEgressInsight = &mesh_proto.ZoneEgressInsight{
		Subscriptions: []*mesh_proto.DiscoverySubscription{{
			Id: "1",
		}},
	}
	ExternalService = &mesh_proto.ExternalService{
		Networking: &mesh_proto.ExternalService_Networking{
			Address: "192.168.0.1",
		},
		Tags: map[string]string{
			mesh_proto.ZoneTag:    "kuma-1",
			mesh_proto.ServiceTag: "backend",
		},
	}
	CircuitBreaker = &mesh_proto.CircuitBreaker{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.CircuitBreaker_Conf{
			Interval: util_proto.Duration(99 * time.Second),
			Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{
				TotalErrors: &mesh_proto.CircuitBreaker_Conf_Detectors_Errors{Consecutive: util_proto.UInt32(3)},
			},
		},
	}
	HealthCheck = &mesh_proto.HealthCheck{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.HealthCheck_Conf{
			Interval:           util_proto.Duration(time.Second * 5),
			Timeout:            util_proto.Duration(time.Second * 7),
			HealthyThreshold:   9,
			UnhealthyThreshold: 11,
		},
	}
	TrafficLog = &mesh_proto.TrafficLog{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.TrafficLog_Conf{
			Backend: "logging-backend",
		},
	}
	TrafficPermission = &mesh_proto.TrafficPermission{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
	}
	TrafficRoute = &mesh_proto.TrafficRoute{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.TrafficRoute_Conf{
			Split: []*mesh_proto.TrafficRoute_Split{{
				Weight: util_proto.UInt32(10),
				Destination: map[string]string{
					mesh_proto.ServiceTag: "*",
				},
			}},
		},
	}
	TrafficTrace = &mesh_proto.TrafficTrace{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{mesh_proto.ServiceTag: "*"},
		}},
		Conf: &mesh_proto.TrafficTrace_Conf{
			Backend: "tracing-backend",
		},
	}
	ProxyTemplate = &mesh_proto.ProxyTemplate{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{mesh_proto.ServiceTag: "*"},
		}},
		Conf: &mesh_proto.ProxyTemplate_Conf{
			Imports: []string{"default-proxy"},
		},
	}
	Retry = &mesh_proto.Retry{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.Retry_Conf{
			Http: &mesh_proto.Retry_Conf_Http{
				NumRetries:    util_proto.UInt32(5),
				PerTryTimeout: util_proto.Duration(time.Second * 200000000),
				BackOff: &mesh_proto.Retry_Conf_BackOff{
					BaseInterval: util_proto.Duration(time.Nanosecond * 200000000),
					MaxInterval:  util_proto.Duration(time.Second * 1),
				},
				RetriableStatusCodes: []uint32{500, 502},
			},
		},
	}
	Timeout = &mesh_proto.Timeout{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.Timeout_Conf{
			ConnectTimeout: util_proto.Duration(time.Second * 5),
			Tcp: &mesh_proto.Timeout_Conf_Tcp{
				IdleTimeout: util_proto.Duration(time.Second * 5),
			},
			Http: &mesh_proto.Timeout_Conf_Http{
				RequestTimeout: util_proto.Duration(time.Second * 5),
				IdleTimeout:    util_proto.Duration(time.Second * 5),
			},
			Grpc: &mesh_proto.Timeout_Conf_Grpc{
				StreamIdleTimeout: util_proto.Duration(time.Second * 5),
				MaxStreamDuration: util_proto.Duration(time.Second * 5),
			},
		},
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
	RateLimit = &mesh_proto.RateLimit{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "*",
			},
		}},
		Conf: &mesh_proto.RateLimit_Conf{
			Http: &mesh_proto.RateLimit_Conf_Http{
				Requests: 100,
				Interval: util_proto.Duration(1 * time.Minute),
			},
		},
	}
	Gateway = &mesh_proto.MeshGateway{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "gateway",
			},
		}},
		Tags: map[string]string{
			"gateway-name": "philip",
		},
		Conf: &mesh_proto.MeshGateway_Conf{
			Listeners: []*mesh_proto.MeshGateway_Listener{{
				Hostname: "philip.example.com",
				Port:     8080,
				Protocol: mesh_proto.MeshGateway_Listener_HTTP,
				Tags: map[string]string{
					"port": "8080",
				},
			}},
		},
	}
	GatewayRoute = &mesh_proto.MeshGatewayRoute{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "gateway",
			},
		}},
		Conf: &mesh_proto.MeshGatewayRoute_Conf{
			Route: &mesh_proto.MeshGatewayRoute_Conf_Http{
				Http: &mesh_proto.MeshGatewayRoute_HttpRoute{},
			},
		},
	}
	VirtualOutbound = &mesh_proto.VirtualOutbound{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{
				mesh_proto.ServiceTag: "virtual-outbound",
			},
		}},
		Conf: &mesh_proto.VirtualOutbound_Conf{
			Host: "{{.service}}.mesh",
			Port: "{{.port}}",
			Parameters: []*mesh_proto.VirtualOutbound_Conf_TemplateParameter{
				{Name: "port"},
				{Name: "service", TagKey: mesh_proto.ServiceTag},
			},
		},
	}
	MeshTrafficPermission = &meshtrafficpermissions.MeshTrafficPermission{
		TargetRef: &common_api.TargetRef{
			Kind: "Mesh",
		},
		From: []meshtrafficpermissions.From{
			{
				TargetRef: common_api.TargetRef{
					Kind: "Mesh",
				},
				Default: meshtrafficpermissions.Conf{
					Action: "Allow",
				},
			},
		},
	}
	MeshAccessLog = &meshaccesslog.MeshAccessLog{
		TargetRef: &common_api.TargetRef{
			Kind: "Mesh",
		},
		From: []meshaccesslog.From{
			{
				TargetRef: common_api.TargetRef{
					Kind: "Mesh",
				},
				Default: meshaccesslog.Conf{
					Backends: &[]meshaccesslog.Backend{
						{
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
