package server

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/Kong/kuma/pkg/kds"

	"github.com/Kong/kuma/pkg/core/resources/apis/system"
	"github.com/Kong/kuma/pkg/core/resources/model"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"
	envoy_server "github.com/envoyproxy/go-control-plane/pkg/server/v2"
	"github.com/go-logr/logr"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesh_proto "github.com/Kong/kuma/api/mesh/v1alpha1"
	mesh_core "github.com/Kong/kuma/pkg/core/resources/apis/mesh"
)

type Server interface {
	mesh_proto.KumaDiscoveryServiceServer
}

func NewServer(config envoy_cache.Cache, callbacks envoy_server.Callbacks, log logr.Logger) Server {
	return &server{cache: config, callbacks: callbacks, log: log}
}

// server is a simplified version of the original XDS server at
// https://github.com/envoyproxy/go-control-plane/blob/master/pkg/server/server.go
type server struct {
	cache     envoy_cache.Cache
	callbacks envoy_server.Callbacks

	// streamCount for counting bi-di streams
	streamCount int64

	log logr.Logger
}

type stream interface {
	grpc.ServerStream

	Send(*envoy.DiscoveryResponse) error
	Recv() (*envoy.DiscoveryRequest, error)
}

// watches for all xDS resource types
type watches struct {
	meshes             chan envoy_cache.Response
	ingresses          chan envoy_cache.Response
	circuitBreakers    chan envoy_cache.Response
	faultInjections    chan envoy_cache.Response
	healthChecks       chan envoy_cache.Response
	trafficLogs        chan envoy_cache.Response
	trafficPermissions chan envoy_cache.Response
	trafficRoutes      chan envoy_cache.Response
	trafficTraces      chan envoy_cache.Response
	proxyTemplates     chan envoy_cache.Response
	secrets            chan envoy_cache.Response

	meshesCancel             func()
	ingressesCancel          func()
	circuitBreakersCancel    func()
	faultInjectionsCancel    func()
	healthChecksCancel       func()
	trafficLogsCancel        func()
	trafficPermissionsCancel func()
	trafficRoutesCancel      func()
	trafficTracesCancel      func()
	proxyTemplatesCancel     func()
	secretsCancel            func()

	meshesNonce             string
	ingressesNonce          string
	circuitBreakersNonce    string
	faultInjectionsNonce    string
	healthChecksNonce       string
	trafficLogsNonce        string
	trafficPermissionsNonce string
	trafficRoutesNonce      string
	trafficTracesNonce      string
	proxyTemplatesNonce     string
	secretsNonce            string
}

// Cancel all watches
func (values watches) Cancel() {
	if values.meshesCancel != nil {
		values.meshesCancel()
	}
	if values.ingressesCancel != nil {
		values.ingressesCancel()
	}
	if values.circuitBreakersCancel != nil {
		values.circuitBreakersCancel()
	}
	if values.faultInjectionsCancel != nil {
		values.faultInjectionsCancel()
	}
	if values.healthChecksCancel != nil {
		values.healthChecksCancel()
	}
	if values.trafficLogsCancel != nil {
		values.trafficLogsCancel()
	}
	if values.trafficPermissionsCancel != nil {
		values.trafficPermissionsCancel()
	}
	if values.trafficRoutesCancel != nil {
		values.trafficRoutesCancel()
	}
	if values.trafficTracesCancel != nil {
		values.trafficTracesCancel()
	}
	if values.proxyTemplatesCancel != nil {
		values.proxyTemplatesCancel()
	}
	if values.secretsCancel != nil {
		values.secretsCancel()
	}
}

func createResponse(resp *envoy_cache.Response, typeURL string) (*envoy.DiscoveryResponse, error) {
	if resp == nil {
		return nil, errors.New("missing response")
	}
	resources := make([]*any.Any, len(resp.Resources))
	for i := 0; i < len(resp.Resources); i++ {
		// Envoy relies on serialized protobuf bytes for detecting changes to the resources.
		// This requires deterministic serialization.
		b := proto.NewBuffer(nil)
		b.SetDeterministic(true)
		err := b.Marshal(resp.Resources[i])
		if err != nil {
			return nil, err
		}
		resources[i] = &any.Any{
			TypeUrl: kds.KumaResource,
			Value:   b.Bytes(),
		}
	}
	out := &envoy.DiscoveryResponse{
		VersionInfo: resp.Version,
		Resources:   resources,
		TypeUrl:     typeURL,
	}
	return out, nil
}

// process handles a bi-di stream request
func (s *server) process(stream stream, reqCh <-chan *envoy.DiscoveryRequest) (err error) {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

	log := s.log.WithValues("streamID", streamID)
	defer func() {
		if err != nil {
			log.Error(err, "xDS stream terminated with an error")
		}
	}()

	// unique nonce generator for req-resp pairs per xDS stream; the server
	// ignores stale nonces. nonce is only modified within send() function.
	var streamNonce int64

	// a collection of watches per request type
	var values watches
	defer func() {
		values.Cancel()
		if s.callbacks != nil {
			s.callbacks.OnStreamClosed(streamID)
		}
	}()

	// sends a response by serializing to protobuf Any
	send := func(resp envoy_cache.Response, resourceType model.ResourceType) (string, error) {
		out, err := createResponse(&resp, string(resourceType))
		if err != nil {
			return "", err
		}

		// increment nonce
		streamNonce++
		out.Nonce = strconv.FormatInt(streamNonce, 10)
		if s.callbacks != nil {
			s.callbacks.OnStreamResponse(streamID, &resp.Request, out)
		}

		return out.Nonce, stream.Send(out)
	}

	if s.callbacks != nil {
		if err := s.callbacks.OnStreamOpen(stream.Context(), streamID, ""); err != nil {
			return err
		}
	}

	// node may only be set on the first discovery request
	var node = &envoy_core.Node{}

	for {
		select {
		case resp, more := <-values.meshes:
			if !more {
				return status.Errorf(codes.Unavailable, "meshes watch failed")
			}
			nonce, err := send(resp, mesh_core.MeshType)
			if err != nil {
				return err
			}
			values.meshesNonce = nonce

		case resp, more := <-values.ingresses:
			if !more {
				return status.Errorf(codes.Unavailable, "ingresses watch failed")
			}
			nonce, err := send(resp, mesh_core.DataplaneType)
			if err != nil {
				return err
			}
			values.ingressesNonce = nonce

		case resp, more := <-values.circuitBreakers:
			if !more {
				return status.Errorf(codes.Unavailable, "circuitBreakers watch failed")
			}
			nonce, err := send(resp, mesh_core.CircuitBreakerType)
			if err != nil {
				return err
			}
			values.circuitBreakersNonce = nonce

		case resp, more := <-values.faultInjections:
			if !more {
				return status.Errorf(codes.Unavailable, "faultInjections watch failed")
			}
			nonce, err := send(resp, mesh_core.FaultInjectionType)
			if err != nil {
				return err
			}
			values.faultInjectionsNonce = nonce

		case resp, more := <-values.healthChecks:
			if !more {
				return status.Errorf(codes.Unavailable, "healthChecks watch failed")
			}
			nonce, err := send(resp, mesh_core.HealthCheckType)
			if err != nil {
				return err
			}
			values.healthChecksNonce = nonce

		case resp, more := <-values.trafficLogs:
			if !more {
				return status.Errorf(codes.Unavailable, "trafficLogs watch failed")
			}
			nonce, err := send(resp, mesh_core.TrafficLogType)
			if err != nil {
				return err
			}
			values.trafficLogsNonce = nonce

		case resp, more := <-values.trafficPermissions:
			if !more {
				return status.Errorf(codes.Unavailable, "trafficPermissions watch failed")
			}
			nonce, err := send(resp, mesh_core.TrafficPermissionType)
			if err != nil {
				return err
			}
			values.trafficPermissionsNonce = nonce

		case resp, more := <-values.trafficRoutes:
			if !more {
				return status.Errorf(codes.Unavailable, "trafficRoutes watch failed")
			}
			nonce, err := send(resp, mesh_core.TrafficRouteType)
			if err != nil {
				return err
			}
			values.trafficRoutesNonce = nonce

		case resp, more := <-values.trafficTraces:
			if !more {
				return status.Errorf(codes.Unavailable, "trafficTraces watch failed")
			}
			nonce, err := send(resp, mesh_core.TrafficTraceType)
			if err != nil {
				return err
			}
			values.trafficTracesNonce = nonce

		case resp, more := <-values.proxyTemplates:
			if !more {
				return status.Errorf(codes.Unavailable, "proxyTemplates watch failed")
			}
			nonce, err := send(resp, mesh_core.ProxyTemplateType)
			if err != nil {
				return err
			}
			values.proxyTemplatesNonce = nonce

		case resp, more := <-values.secrets:
			if !more {
				return status.Errorf(codes.Unavailable, "secrets watch failed")
			}
			nonce, err := send(resp, system.SecretType)
			if err != nil {
				return err
			}
			values.secretsNonce = nonce

		case req, more := <-reqCh:
			// input stream ended or errored out
			if !more {
				return nil
			}
			if req == nil {
				return status.Errorf(codes.Unavailable, "empty request")
			}

			// node field in discovery request is delta-compressed
			if req.Node != nil {
				node = req.Node
			} else {
				req.Node = node
			}

			// nonces can be reused across streams; we verify nonce only if nonce is not initialized
			nonce := req.GetResponseNonce()

			// type URL is required for ADS but is implicit for xDS
			if req.TypeUrl == "" {
				return status.Errorf(codes.InvalidArgument, "type URL is required for KDS")
			}

			if s.callbacks != nil {
				if err := s.callbacks.OnStreamRequest(streamID, req); err != nil {
					return err
				}
			}

			requestResourceType := model.ResourceType(req.TypeUrl)
			// cancel existing watches to (re-)request a newer version
			switch {
			case requestResourceType == mesh_core.MeshType && (values.meshesNonce == "" || values.meshesNonce == nonce):
				if values.meshesCancel != nil {
					values.meshesCancel()
				}
				values.meshes, values.meshesCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.DataplaneType && (values.ingressesNonce == "" || values.ingressesNonce == nonce):
				if values.ingressesCancel != nil {
					values.ingressesCancel()
				}
				values.ingresses, values.ingressesCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.CircuitBreakerType && (values.circuitBreakersNonce == "" || values.circuitBreakersNonce == nonce):
				if values.circuitBreakersCancel != nil {
					values.circuitBreakersCancel()
				}
				values.circuitBreakers, values.circuitBreakersCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.FaultInjectionType && (values.faultInjectionsNonce == "" || values.faultInjectionsNonce == nonce):
				if values.faultInjectionsCancel != nil {
					values.faultInjectionsCancel()
				}
				values.faultInjections, values.faultInjectionsCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.HealthCheckType && (values.healthChecksNonce == "" || values.healthChecksNonce == nonce):
				if values.healthChecksCancel != nil {
					values.healthChecksCancel()
				}
				values.healthChecks, values.healthChecksCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.TrafficLogType && (values.trafficLogsNonce == "" || values.trafficLogsNonce == nonce):
				if values.trafficLogsCancel != nil {
					values.trafficLogsCancel()
				}
				values.trafficLogs, values.trafficLogsCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.TrafficPermissionType && (values.trafficPermissionsNonce == "" || values.trafficPermissionsNonce == nonce):
				if values.trafficPermissionsCancel != nil {
					values.trafficPermissionsCancel()
				}
				values.trafficPermissions, values.trafficPermissionsCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.TrafficRouteType && (values.trafficRoutesNonce == "" || values.trafficRoutesNonce == nonce):
				if values.trafficRoutesCancel != nil {
					values.trafficRoutesCancel()
				}
				values.trafficRoutes, values.trafficRoutesCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.TrafficTraceType && (values.trafficTracesNonce == "" || values.trafficTracesNonce == nonce):
				if values.trafficTracesCancel != nil {
					values.trafficTracesCancel()
				}
				values.trafficTraces, values.trafficTracesCancel = s.cache.CreateWatch(*req)
			case requestResourceType == mesh_core.ProxyTemplateType && (values.proxyTemplatesNonce == "" || values.proxyTemplatesNonce == nonce):
				if values.proxyTemplatesCancel != nil {
					values.proxyTemplatesCancel()
				}
				values.proxyTemplates, values.proxyTemplatesCancel = s.cache.CreateWatch(*req)
			case requestResourceType == system.SecretType && (values.secretsNonce == "" || values.secretsNonce == nonce):
				if values.secretsCancel != nil {
					values.secretsCancel()
				}
				values.secrets, values.secretsCancel = s.cache.CreateWatch(*req)
			}
		}
	}
}

// handler converts a blocking read call to channels and initiates stream processing
func (s *server) handler(stream stream) error {
	// a channel for receiving incoming requests
	reqCh := make(chan *envoy.DiscoveryRequest)
	reqStop := int32(0)
	go func() {
		for {
			req, err := stream.Recv()
			if atomic.LoadInt32(&reqStop) != 0 {
				return
			}
			if err != nil {
				close(reqCh)
				return
			}
			reqCh <- req
		}
	}()

	err := s.process(stream, reqCh)

	atomic.StoreInt32(&reqStop, 1)

	return err
}

func (s *server) StreamKumaResources(stream mesh_proto.KumaDiscoveryService_StreamKumaResourcesServer) error {
	return s.handler(stream)
}

func (s *server) DeltaKumaResources(_ mesh_proto.KumaDiscoveryService_DeltaKumaResourcesServer) error {
	return errors.New("not implemented")
}

func (s *server) FetchKumaResources(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	if req == nil {
		return nil, status.Errorf(codes.Unavailable, "empty request")
	}
	return s.Fetch(ctx, req)
}

// Fetch is the universal fetch method.
func (s *server) Fetch(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	if s.callbacks != nil {
		if err := s.callbacks.OnFetchRequest(ctx, req); err != nil {
			return nil, err
		}
	}
	resp, err := s.cache.Fetch(ctx, *req)
	if err != nil {
		return nil, err
	}
	out, err := createResponse(resp, req.TypeUrl)
	if s.callbacks != nil {
		s.callbacks.OnFetchResponse(req, out)
	}
	return out, err
}
