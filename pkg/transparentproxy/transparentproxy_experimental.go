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

	defaultConfig := kumanet_config.DefaultConfig()

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
