package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/kumahq/kuma-net/iptables/builder"
	"github.com/kumahq/kuma-net/iptables/config"
)

func convertToUint16(field string, value string) uint16 {
	converted, err := strconv.ParseUint(value, 10, 16)
	if err != nil {
		log.Error(err, "failed to convert to int16", zap.String(field, value))
		os.Exit(1)
	}
	return uint16(converted)
}

func convertCommaSeparatedString(list string) []uint16 {
	splitted := strings.Split(list, ",")
	mapped := make([]uint16, len(splitted))

	for i, value := range splitted {
		mapped[i] = convertToUint16(strconv.Itoa(i), value)
	}

	return mapped
}

func Inject(netns string, intermediateConfig *IntermediateConfig) error {
	cfg, err := mapToConfig(intermediateConfig)
	if err != nil {
		return err
	}

	namespace, err := ns.GetNS(netns)
	if err != nil {
		err = fmt.Errorf("failed to open namespace %q: %s", namespace, err)
		return err
	}
	defer namespace.Close()

	if err = namespace.Do(func(_ ns.NetNS) error {
		_, err := builder.RestoreIPTables(cfg)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func mapToConfig(intermediateConfig *IntermediateConfig) (config.Config, error) {
	cfg := config.Config{
		RuntimeOutput: ioutil.Discard,
		Owner: config.Owner{
			UID: intermediateConfig.noRedirectUID,
		},
		Redirect: config.Redirect{
			Outbound: config.TrafficFlow{
				Port:         convertToUint16("inbound port", intermediateConfig.targetPort),
				ExcludePorts: convertCommaSeparatedString(intermediateConfig.excludeOutboundPorts),
			},
		},
	}

	isGateway, err := GetEnabled(intermediateConfig.isGateway)
	if err != nil {
		return config.Config{}, err
	}
	redirectInbound := !isGateway
	if redirectInbound {
		cfg.Redirect.Inbound = config.TrafficFlow{
			Port:         convertToUint16("inbound port", intermediateConfig.inboundPort),
			PortIPv6:     convertToUint16("inbount port ipv6", intermediateConfig.inboundPortV6),
			ExcludePorts: convertCommaSeparatedString(intermediateConfig.excludeInboundPorts),
		}
	}

	useBuiltinDNS, err := GetEnabled(intermediateConfig.builtinDNS)
	if err != nil {
		return config.Config{}, err
	}
	if useBuiltinDNS {
		cfg.Redirect.DNS = config.DNS{
			Enabled: true,
			Port:    convertToUint16("buildin dns port", intermediateConfig.builtinDNSPort),
		}
	}
	return cfg, nil
}

func GetEnabled(value string) (bool, error) {
	switch strings.ToLower(value) {
	case "enabled", "true":
		return true, nil
	case "disabled", "false":
		return false, nil
	default:
		return false, errors.Errorf("wrong value \"%s\", available values are: \"enabled\", \"disabled\"", value)
	}
}
