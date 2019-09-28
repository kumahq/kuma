package accesslogs


import (
	"fmt"
	"github.com/Kong/kuma/pkg/core"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	"github.com/envoyproxy/go-control-plane/envoy/data/accesslog/v2"
	v2 "github.com/envoyproxy/go-control-plane/envoy/service/accesslog/v2"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"strings"
	"time"
)

var logger = core.Log.WithName("accesslogs")

type accessLogServer struct {
	server *grpc.Server
	conn   net.Conn
}

func NewAccessLogServer() *accessLogServer {
	return &accessLogServer{
		server: grpc.NewServer(),
	}
}

func (a *accessLogServer) StreamAccessLogs(server v2.AccessLogService_StreamAccessLogsServer) error {
	msg, err := server.Recv()
	if err != nil {
		return err
	}

	for _, entry := range msg.LogEntries.(*v2.StreamAccessLogsMessage_HttpLogs).HttpLogs.LogEntry {
		split := strings.Split(msg.GetIdentifier().GetLogName(), ";")
		address := split[0]
		format := split[1]
		println(entry.String())
		entry := logEntry(entry, format)
		if err := a.sendLog(address, entry); err != nil {
			return errors.Wrap(err, "could not send log")
		}
	}
	return nil
}

func logEntry(entry *envoy_data_accesslog_v2.HTTPAccessLogEntry, format string) string {
	addrToString := func(addr *envoy_core.Address) string {
		return fmt.Sprintf("%s:%d", addr.GetSocketAddress().GetAddress(), addr.GetSocketAddress().GetPortValue())
	}
	placeholders := map[string]string{
		"%START_TIME%":                entry.GetCommonProperties().GetStartTime().Format(time.RFC3339),
		"%DOWNSTREAM_REMOTE_ADDRESS%": addrToString(entry.GetCommonProperties().GetDownstreamRemoteAddress()),
		"%DOWNSTREAM_LOCAL_ADDRESS%":  addrToString(entry.GetCommonProperties().GetDownstreamLocalAddress()),
		"%UPSTREAM_LOCAL_ADDRESS%":    addrToString(entry.GetCommonProperties().GetUpstreamLocalAddress()),
		"%UPSTREAM_CLUSTER%":          entry.GetCommonProperties().GetUpstreamCluster(),
		"%BYTES_RECEIVED%":            strconv.Itoa(int(entry.GetResponse().GetResponseBodyBytes())),
		"%BYTES_SENT%":                strconv.Itoa(int(entry.GetRequest().GetRequestBodyBytes())),
	}
	log := format
	for placeholder, value := range placeholders {
		log = strings.ReplaceAll(log, placeholder, value)
	}
	return log
}

func (a *accessLogServer) sendLog(address string, log string) error {
	if a.conn == nil {
		if err := a.connect(address); err != nil {
			return errors.Wrapf(err, "could not connect to service %s", address)
		}
	}
	if a.conn.RemoteAddr().String() != address {
		// try to close old connection when address changes
		_ = a.conn.Close()
		if err := a.connect(address); err != nil {
			return errors.Wrapf(err, "could not connect to service %s", address)
		}
	}
	if _, err := a.conn.Write([]byte(log)); err != nil {
		// retry connection and sending on error once
		if err := a.connect(address); err != nil {
			return errors.Wrapf(err, "could not connect to service %s", address)
		}
		if _, err := a.conn.Write([]byte(log)); err != nil {
			return err
		}
	}
	return nil
}

func (a *accessLogServer) connect(address string) error {
	logger.Info("Connecting to TCP service", "address", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}
	logger.Info("Connected", "address", address)
	a.conn = conn
	return nil
}

func (a *accessLogServer) Start(port uint32) error {
	v2.RegisterAccessLogServiceServer(a.server, a)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	if err := a.server.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (a *accessLogServer) Close() {
	a.server.GracefulStop()
}

var _ v2.AccessLogServiceServer = &accessLogServer{}
