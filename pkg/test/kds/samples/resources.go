package samples

import (
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/wrappers"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
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
				"service": "*",
				"tag0":    "version0",
				"tag1":    "version1",
				"tag2":    "version2",
				"tag3":    "version3",
				"tag4":    "version4",
				"tag5":    "version5",
				"tag6":    "version6",
				"tag7":    "version7",
				"tag8":    "version8",
				"tag9":    "version9",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.FaultInjection_Conf{
			Abort: &mesh_proto.FaultInjection_Conf_Abort{
				Percentage: &wrappers.DoubleValue{Value: 90},
				HttpStatus: &wrappers.UInt32Value{Value: 404},
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
	DataplaneInsight = &mesh_proto.DataplaneInsight{
		MTLS: &mesh_proto.DataplaneInsight_MTLS{
			CertificateRegenerations: 3,
		},
	}
	Ingress = &mesh_proto.Dataplane{
		Networking: &mesh_proto.Dataplane_Networking{
			Ingress: &mesh_proto.Dataplane_Networking_Ingress{
				AvailableServices: []*mesh_proto.Dataplane_Networking_Ingress_AvailableService{{
					Tags: map[string]string{
						"service": "backend",
					}},
				},
			},
			Address: "192.168.0.1",
		},
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
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.CircuitBreaker_Conf{
			Detectors: &mesh_proto.CircuitBreaker_Conf_Detectors{},
		},
	}
	HealthCheck = &mesh_proto.HealthCheck{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.HealthCheck_Conf{
			Interval: &duration.Duration{Seconds: 5},
			Timeout:  &duration.Duration{Seconds: 7},
		},
	}
	TrafficLog = &mesh_proto.TrafficLog{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.TrafficLog_Conf{
			Backend: "logging-backend",
		},
	}
	TrafficPermission = &mesh_proto.TrafficPermission{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"kuma.io/service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"kuma.io/service": "*",
			},
		}},
	}
	TrafficRoute = &mesh_proto.TrafficRoute{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.TrafficRoute_Conf{
			Split: []*mesh_proto.TrafficRoute_Split{{
				Weight: 10,
				Destination: map[string]string{
					"version": "v2",
				},
			}},
		},
	}
	TrafficTrace = &mesh_proto.TrafficTrace{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{"serivce": "*"},
		}},
		Conf: &mesh_proto.TrafficTrace_Conf{
			Backend: "tracing-backend",
		},
	}
	ProxyTemplate = &mesh_proto.ProxyTemplate{
		Selectors: []*mesh_proto.Selector{{
			Match: map[string]string{"serivce": "*"},
		}},
		Conf: &mesh_proto.ProxyTemplate_Conf{
			Imports: []string{"default-kuma-profile"},
		},
	}
	Retry = &mesh_proto.Retry{
		Sources: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Destinations: []*mesh_proto.Selector{{
			Match: map[string]string{
				"service": "*",
			},
		}},
		Conf: &mesh_proto.Retry_Conf{
			Http: &mesh_proto.Retry_Conf_Http{
				NumRetries: &wrappers.UInt32Value{
					Value: 5,
				},
				PerTryTimeout: &duration.Duration{
					Seconds: 200000000,
				},
				BackOff: &mesh_proto.Retry_Conf_BackOff{
					BaseInterval: &duration.Duration{
						Nanos: 200000000,
					},
					MaxInterval: &duration.Duration{
						Seconds: 1,
					},
				},
				RetriableStatusCodes: []uint32{500, 502},
			},
		},
	}
	Secret = &system_proto.Secret{
		Data: &wrappers.BytesValue{Value: []byte("secret key")},
	}
	Config = &system_proto.Config{
		Config: "sample config",
	}
)
