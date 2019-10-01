package accesslogs

import (
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	kumadp "github.com/Kong/kuma/pkg/config/app/kuma-dp"
	"github.com/Kong/kuma/pkg/core"

	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_data_accesslog_v2 "github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
	v2 "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
)

var logger = core.Log.WithName("accesslogs-server")

const (
	defaultConnectTimeout = 5 * time.Second
)

type accessLogServer struct {
	server *grpc.Server

	// streamCount for counting streams
	streamCount int64
}

func NewAccessLogServer() *accessLogServer {
	return &accessLogServer{
		server: grpc.NewServer(),
	}
}

func (s *accessLogServer) StreamAccessLogs(stream v2.AccessLogService_StreamAccessLogsServer) (err error) {
	// increment stream count
	streamID := atomic.AddInt64(&s.streamCount, 1)

	totalRequests := 0

	log := logger.WithValues("streamID", streamID)
	log.Info("starting a new Access Logs stream")
	defer func() {
		log.Info("Access Logs stream is terminated", "totalRequests", totalRequests)
		if err != nil {
			log.Error(err, "Access Logs stream terminated with an error")
		}
	}()

	initialized := false
	var address, format string
	var conn net.Conn
	for {
		msg, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		totalRequests++

		if !initialized {
			initialized = true

			parts := strings.SplitN(msg.Identifier.GetLogName(), ";", 2)
			if len(parts) != 2 {
				return errors.Errorf("failed to initialize Access Logs stream: invalid log name %q: expected %d components, got %d", msg.Identifier.GetLogName(), 2, len(parts))
			}
			address = parts[0]
			format = parts[1]
			conn, err = s.connect(address, log)
			if err != nil {
				return err
			}
			defer conn.Close()
		}

		var httpLogs []*envoy_data_accesslog_v2.HTTPAccessLogEntry
		switch logEntries := msg.LogEntries.(type) {
		case *v2.StreamAccessLogsMessage_HttpLogs:
			httpLogs = logEntries.HttpLogs.GetLogEntry()
		case *v2.StreamAccessLogsMessage_TcpLogs:
			return errors.New("TcpLogs entries are not supported yet")
		default:
			return errors.Errorf("unknown type of log entries: %T", msg.LogEntries)
		}

		for _, httpLogEntry := range httpLogs {
			entry := formatEntry(httpLogEntry, format)
			if err := s.sendLog(conn, entry); err != nil {
				return errors.Wrap(err, "could not send log entry to a TCP logging backend")
			}
		}
	}
}

func formatEntry(entry *envoy_data_accesslog_v2.HTTPAccessLogEntry, format string) string {
	addrToString := func(addr *envoy_core.Address) string {
		return fmt.Sprintf("%s:%d", addr.GetSocketAddress().GetAddress(), addr.GetSocketAddress().GetPortValue())
	}
	connectionTime := int64(0)
	if entry.GetCommonProperties().GetTimeToLastDownstreamTxByte() != nil {
		connectionTime = int64(*entry.GetCommonProperties().GetTimeToLastDownstreamTxByte() / time.Millisecond)
	}
	placeholders := map[string]string{
		"%START_TIME%":                entry.GetCommonProperties().GetStartTime().Format(time.RFC3339),
		"%DOWNSTREAM_REMOTE_ADDRESS%": addrToString(entry.GetCommonProperties().GetDownstreamRemoteAddress()),
		"%DOWNSTREAM_LOCAL_ADDRESS%":  addrToString(entry.GetCommonProperties().GetDownstreamLocalAddress()),
		"%UPSTREAM_HOST%":             addrToString(entry.GetCommonProperties().GetUpstreamRemoteAddress()),
		"%UPSTREAM_REMOTE_ADDRESS%":   addrToString(entry.GetCommonProperties().GetUpstreamRemoteAddress()),
		"%UPSTREAM_LOCAL_ADDRESS%":    addrToString(entry.GetCommonProperties().GetUpstreamLocalAddress()),
		"%UPSTREAM_CLUSTER%":          entry.GetCommonProperties().GetUpstreamCluster(),
		"%BYTES_RECEIVED%":            strconv.FormatUint(entry.GetResponse().GetResponseBodyBytes(), 10),
		"%BYTES_SENT%":                strconv.FormatUint(entry.GetRequest().GetRequestBodyBytes(), 10),
		"%DURATION%":                  strconv.FormatInt(connectionTime, 10),
		"%RESPONSE_TX_DURATION%":      strconv.FormatInt(connectionTime, 10),
	}
	log := format
	for placeholder, value := range placeholders {
		log = strings.ReplaceAll(log, placeholder, value)
	}
	return log
}

func (s *accessLogServer) sendLog(conn net.Conn, log string) error {
	_, err := conn.Write([]byte(log))
	return err
}

func (s *accessLogServer) connect(address string, log logr.Logger) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", address, defaultConnectTimeout)
	if err != nil {
		log.Error(err, "failed to connect to TCP logging backend", "address", address)
		return nil, err
	}
	log.Info("connected to TCP logging backend", "address", address)
	return conn, nil
}

func (s *accessLogServer) Start(dataplane kumadp.Dataplane) error {
	v2.RegisterAccessLogServiceServer(s.server, s)
	address := fmt.Sprintf("/tmp/kuma-access-logs-%s-%s.sock", dataplane.Name, dataplane.Mesh)
	lis, err := net.Listen("unix", address)
	if err != nil {
		return err
	}
	logger.Info("starting", "address", fmt.Sprintf("unix://%s", address))
	if err := s.server.Serve(lis); err != nil {
		logger.Error(err, "terminated with an error")
		return err
	}
	return nil
}

func (s *accessLogServer) Close() {
	s.server.GracefulStop()
}

var _ v2.AccessLogServiceServer = &accessLogServer{}
