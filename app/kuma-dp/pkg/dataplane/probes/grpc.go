package probes

import (
	"context"
	"fmt"
	grpchealth "github.com/kumahq/kuma/app/kuma-dp/pkg/dataplane/probes/api"
	"github.com/kumahq/kuma/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
)

func (p *Prober) probeGRPC(writer http.ResponseWriter, req *http.Request) {
	// /grpc/<port>/<service>

	opts := []grpc.DialOption{
		grpc.WithUserAgent(fmt.Sprintf("kube-probe/%s", version.Build.Version)),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return createProbeDialer().DialContext(ctx, "tcp", addr)
		}),
	}

	ctx, cancel := context.WithTimeout(context.Background(), getTimeout(req))

	defer cancel()

	port, err := getPort(req, tcpGRPCPathPattern)
	if err != nil {
		logger.V(1).Info("invalid port number", "error", err)
		writeProbeResult(writer, Unkown)
		return
	}

	addr := net.JoinHostPort(p.IPAddress, fmt.Sprintf("%d", port))
	conn, err := grpc.DialContext(ctx, addr, opts...)

	if err != nil {
		logger.V(1).Info("unable to connect to upstream server", "error", err)
		writeProbeResult(writer, Unhealthy)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	client := grpchealth.NewHealthClient(conn)
	resp, err := client.Check(metadata.NewOutgoingContext(ctx, make(metadata.MD)), &grpchealth.HealthCheckRequest{
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
		logger.V(1).Info(fmt.Sprintf("the upstream server returned status %s", resp.GetStatus()))
		writeProbeResult(writer, Unhealthy)
		return
	}

	writeProbeResult(writer, Healthy)
	return
}
