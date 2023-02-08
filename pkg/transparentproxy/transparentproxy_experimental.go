package transparentproxy

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma-net/iptables/builder"
	kumanet_config "github.com/kumahq/kuma-net/iptables/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/config"
)

var _ TransparentProxy = &ExperimentalTransparentProxy{}

type ExperimentalTransparentProxy struct{}

<<<<<<< HEAD
=======
func hasLocalIPv6() (bool, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return false, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() &&
			ipnet.IP.To4() == nil {
			return true, nil
		}
	}

	return false, nil
}

// ShouldEnableIPv6 checks if system supports IPv6. The port has a value of
// RedirectPortInBoundV6 and when equals 0 means that IPv6 was disabled by the user.
func ShouldEnableIPv6(port uint16) (bool, error) {
	if port == 0 {
		return false, nil
	}

	hasIPv6Address, err := hasLocalIPv6()
	if !hasIPv6Address || err != nil {
		return false, err
	}

	// We are executing this command to work around the problem with COS_CONTAINERD
	// image which is being used on GKE nodes. This image is missing "ip6tables_nat"
	// kernel module which is adding `nat` table, so we are checking if this table
	// exists and if so, we are assuming we can safely proceed with ip6tables
	// ref. https://github.com/kumahq/kuma/issues/2046
	err = exec.Command(constants.IP6TABLES, "-t", constants.NAT, "-L").Run()

	return err == nil, nil
}

func parseUint16(port string) (uint16, error) {
	parsedPort, err := strconv.ParseUint(port, 10, 16)
	if err != nil {
		return 0, fmt.Errorf("value %s, is not valid uint16", port)
	}

	return uint16(parsedPort), nil
}

>>>>>>> 134794214 (fix(tproxy): fix disabling ipv6 for tproxy (#5923))
func splitPorts(ports string) ([]uint16, error) {
	var result []uint16

	for _, port := range strings.Split(ports, ",") {
		p, err := strconv.ParseUint(port, 10, 16)
		if err != nil {
			return nil, errors.Wrapf(err, "port (%s), is not valid uint16", port)
		}

		result = append(result, uint16(p))
	}

	return result, nil
}

func (tp *ExperimentalTransparentProxy) Setup(tpConfig *config.TransparentProxyConfig) (string, error) {
	redirectInboundPort, err := strconv.ParseUint(tpConfig.RedirectPortInBound, 10, 16)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"inbound redirect port (%s), is not valid uint16",
			tpConfig.RedirectPortInBound,
		)
	}

	redirectOutboundPort, err := strconv.ParseUint(tpConfig.RedirectPortOutBound, 10, 16)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"outbound redirect port (%s), is not valid uint16",
			tpConfig.RedirectPortOutBound,
		)
	}

	agentDNSListenerPort, err := strconv.ParseUint(tpConfig.AgentDNSListenerPort, 10, 16)
	if err != nil {
		return "", errors.Wrapf(
			err,
			"outbound redirect port (%s), is not valid uint16",
			tpConfig.RedirectPortOutBound,
		)
	}

	var excludeInboundPorts []uint16
	if tpConfig.ExcludeInboundPorts != "" {
		excludeInboundPorts, err = splitPorts(tpConfig.ExcludeInboundPorts)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse inbound ports to exclude")
		}
	}

	var excludeOutboundPorts []uint16
	if tpConfig.ExcludeOutboundPorts != "" {
		excludeOutboundPorts, err = splitPorts(tpConfig.ExcludeOutboundPorts)
		if err != nil {
			return "", errors.Wrap(err, "cannot parse outbound ports to exclude")
		}
	}

<<<<<<< HEAD
	defaultConfig := kumanet_config.DefaultConfig()
=======
	ipv6, err := ShouldEnableIPv6(redirectInboundPortIPv6)
	if err != nil {
		return "", errors.Wrap(err, "cannot verify if IPv6 should be enabled")
	}
>>>>>>> 134794214 (fix(tproxy): fix disabling ipv6 for tproxy (#5923))

	cfg := &kumanet_config.Config{
		Owner: &kumanet_config.Owner{
			UID: tpConfig.UID,
			GID: tpConfig.GID,
		},
		Redirect: &kumanet_config.Redirect{
			NamePrefix: "KUMA_",
			Inbound: &kumanet_config.TrafficFlow{
				Port:          uint16(redirectInboundPort),
				Chain:         defaultConfig.Redirect.Inbound.Chain,
				RedirectChain: defaultConfig.Redirect.Inbound.RedirectChain,
				ExcludePorts:  excludeInboundPorts,
			},
			Outbound: &kumanet_config.TrafficFlow{
				Port:          uint16(redirectOutboundPort),
				Chain:         defaultConfig.Redirect.Outbound.Chain,
				RedirectChain: defaultConfig.Redirect.Outbound.RedirectChain,
				ExcludePorts:  excludeOutboundPorts,
			},
			DNS: &kumanet_config.DNS{
				Enabled:            tpConfig.RedirectAllDNSTraffic,
				Port:               uint16(agentDNSListenerPort),
				ConntrackZoneSplit: tpConfig.SkipDNSConntrackZoneSplit,
			},
		},
		Verbose: tpConfig.Verbose,
	}

	if tpConfig.DryRun {
		return builder.BuildIPTables(cfg)
	}

	return builder.RestoreIPTables(cfg)
}

func (tp *ExperimentalTransparentProxy) Cleanup(dryRun, verbose bool) (string, error) {
	// TODO implement me
	panic("implement me")
}
