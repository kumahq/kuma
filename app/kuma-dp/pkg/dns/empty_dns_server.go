package dns

import (
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core"
	"github.com/kumahq/kuma/pkg/core/runtime/component"
)

var log = core.Log.WithName("dns")

type EmptyDNSServer struct {
}

func (e *EmptyDNSServer) Start(stop <-chan struct{}) error {
	server := &dns.Server{
		Addr: "127.0.0.1:5691",
		Net:  "udp",
	}

	errChan := make(chan error)
	go func() {
		defer close(errChan)

		err := server.ListenAndServe()
		if err != nil {
			errString := "failed to start the DNS listener."
			if strings.Contains(err.Error(), "bind") {
				errString = bindError("127.0.0.1:5691")
			}
			log.Error(err, errString)
			errChan <- errors.Wrap(err, errString)
		}
	}()

	log.Info("starting", "address", "127.0.0.1:5691")
	select {
	case <-stop:
		log.Info("shutting down the DNS Server")
		return server.Shutdown()
	case err := <-errChan:
		return err
	}
}

func bindError(address string) string {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Sprintf("invalid DNS bind address %s", address)
	}
	return fmt.Sprintf(
		"unable to bind the DNS server to %s.\n\nPlease consider setting KUMA_DNS_SERVER_PORT=5653 (the default).\n"+
			"Then redirect the incoming UDP traffinc on port 53 to it. The `iptables` command for this would be:\n\n"+
			"iptables -t nat -A OUTPUT -p udp -d %s --dport 53 -j DNAT --to-destination %s:5653\n\n"+
			"On hosts which use firewalld, the command would be:\n\n"+
			"firewall-cmd --direct --add-rule ipv4 nat OUTPUT 1 -p udp -d %s --dport 53 -j DNAT --to-destination %s:5653\n\n",
		address,
		host, host,
		host, host)
}


func (e *EmptyDNSServer) NeedLeaderElection() bool {
	return false
}

var _ component.Component = &EmptyDNSServer{}
