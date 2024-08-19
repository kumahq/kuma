package probes

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	grpchealth "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/kumahq/kuma/pkg/version"
)

func (p *Prober) probeGRPC(writer http.ResponseWriter, req *http.Request) {
	// /grpc/<port>

	opts := []grpc.DialOption{
		grpc.WithUserAgent(fmt.Sprintf("kube-probe/%s", version.Build.Version)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return createProbeDialer(p.isPodAddrIPV6).DialContext(ctx, "tcp", addr)
		}),
	}

	port, err := getPort(req, tcpGRPCPathPattern)
	if err != nil {
		logger.V(1).Info("invalid port number", "error", err)
		writeProbeResult(writer, Unknown)
		return
	}

	addr := net.JoinHostPort(p.podAddress, fmt.Sprintf("%d", port))
	conn, err := grpc.NewClient(addr, opts...)
	if err != nil {
		logger.V(1).Info("unable to connect to upstream server", "error", err)
		writeProbeResult(writer, Unhealthy)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	client := grpchealth.NewHealthClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), getTimeout(req))
	defer cancel()
	resp, err := client.Check(metadata.NewOutgoingContext(ctx, make(metadata.MD)), // nolint: contextcheck
		&grpchealth.HealthCheckRequest{
			Service: getGRPCService(req),
		})
	if err != nil {
		stat, ok := status.FromError(err)
		if ok {
			switch stat.Code() {
			case codes.Unimplemented:
				logger.V(1).Info("the upstream server does not implement the grpc health protocol (grpc.health.v1.Health)")
				writeProbeResult(writer, Unhealthy)
				return
			case codes.DeadlineExceeded:
				logger.V(1).Info("the upstream check request did not complete within the timeout")
				writeProbeResult(writer, Unhealthy)
				return
			default:
				logger.V(1).Info(fmt.Sprintf("the upstream check request failed with code %s", stat.Code().String()))
				writeProbeResult(writer, Unhealthy)
				return
			}
		} else {
			logger.V(1).Info("the upstream check request failed")
			writeProbeResult(writer, Unhealthy)
			return
		}
	}

	if resp.GetStatus() != grpchealth.HealthCheckResponse_SERVING {
		logger.V(1).Info("the upstream server returned as not serving", "status", resp.GetStatus())
		writeProbeResult(writer, Unhealthy)
		return
	}

	writeProbeResult(writer, Healthy)
}
