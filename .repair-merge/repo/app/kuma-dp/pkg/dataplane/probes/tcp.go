package probes

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
)

func (p *Prober) probeTCP(writer http.ResponseWriter, req *http.Request) {
	// /grpc/<port>
	port, err := getPort(req, tcpGRPCPathPattern)
	if err != nil {
		logger.V(1).Info("invalid port number", "error", err)
		writeProbeResult(writer, Unknown)
		return
	}

	d := createProbeDialer(p.isPodAddrIPV6)
	d.Timeout = getTimeout(req)
	hostPort := net.JoinHostPort(p.podAddress, strconv.Itoa(port))
	protocol := "tcp"
	if p.isPodAddrIPV6 {
		protocol = "tcp6"
	}
	conn, err := d.Dial(protocol, hostPort)
	if err != nil {
		logger.V(1).Info(fmt.Sprintf("unable to establish TCP connection to %s", hostPort), "error", err)
		writeProbeResult(writer, Unhealthy)
		return
	}

	err = conn.Close()
	if err != nil {
		logger.V(1).Info("unable to close TCP socket", "error", err)
	}
	writeProbeResult(writer, Healthy)
}
