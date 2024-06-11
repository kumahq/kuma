package probes

import (
	"net"
	"net/http"
	"strconv"
)

func (p *Prober) probeTCP(writer http.ResponseWriter, req *http.Request) {
	port, err := getPort(req, tcpGRPCPathPattern)
	if err != nil {
		logger.V(1).Info("invalid port number", "error", err)
		writeProbeResult(writer, Unkown)
		return
	}

	d := createProbeDialer()
	d.Timeout = getTimeout(req)
	hostPort := net.JoinHostPort(p.IPAddress, strconv.Itoa(port))
	conn, err := d.Dial("tcp", hostPort)
	if err != nil {
		writeProbeResult(writer, Unhealthy)
		return
	}

	err = conn.Close()
	if err != nil {
		logger.V(1).Info("unable to close TCP socket", "error", err)
	}
	writeProbeResult(writer, Healthy)
}
