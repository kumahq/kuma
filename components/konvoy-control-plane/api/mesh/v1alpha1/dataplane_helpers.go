package v1alpha1

import (
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

const (
	ServiceTag = "service"
)

type InboundInterface struct {
	DataplaneIP   string
	DataplanePort uint32
	WorkloadPort  uint32
}

func (i InboundInterface) String() string {
	return strings.Join([]string{
		i.DataplaneIP,
		strconv.FormatUint(uint64(i.DataplanePort), 10),
		strconv.FormatUint(uint64(i.WorkloadPort), 10),
	}, ":")
}

const inboundInterfacePattern = `^(?P<dataplane_ip>(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)):(?P<dataplane_port>[0-9]{1,5}):(?P<workload_port>[0-9]{1,5})$`

var inboundInterfaceRegexp = regexp.MustCompile(inboundInterfacePattern)

func ParseInboundInterface(text string) (InboundInterface, error) {
	groups := inboundInterfaceRegexp.FindStringSubmatch(text)
	if groups == nil {
		return InboundInterface{}, errors.Errorf("invalid format: expected %s, got %q", inboundInterfacePattern, text)
	}
	dataplaneIP, err := ParseIP(groups[1])
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid <DATAPLANE_IP> in %q", text)
	}
	dataplanePort, err := ParsePort(groups[2])
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid <DATAPLANE_PORT> in %q", text)
	}
	workloadPort, err := ParsePort(groups[3])
	if err != nil {
		return InboundInterface{}, errors.Wrapf(err, "invalid <WORKLOAD_PORT> in %q", text)
	}
	return InboundInterface{
		DataplaneIP:   dataplaneIP,
		DataplanePort: dataplanePort,
		WorkloadPort:  workloadPort,
	}, nil
}

func ParsePort(text string) (uint32, error) {
	port, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, errors.Wrapf(err, "%q is not a valid port number", text)
	}
	if port < 1 || 65535 < port {
		return 0, errors.Errorf("port number must be in the range [1, 65535] but got %d", port)
	}
	return uint32(port), nil
}

func ParseIP(text string) (string, error) {
	if net.ParseIP(text) == nil {
		return "", errors.Errorf("%q is not a valid IP address", text)
	}
	return text, nil
}

func (n *Dataplane_Networking) GetInboundInterfaces() ([]InboundInterface, error) {
	if n == nil {
		return nil, nil
	}
	ifaces := make([]InboundInterface, len(n.Inbound))
	for i, inbound := range n.Inbound {
		iface, err := ParseInboundInterface(inbound.Interface)
		if err != nil {
			return nil, err
		}
		ifaces[i] = iface
	}
	return ifaces, nil
}

func (d *Dataplane) MatchTags(tags map[string]string) bool {
	if len(tags) == 0 {
		return true
	}
	if d.Networking == nil {
		return false
	}
	for _, inbound := range d.Networking.Inbound {
		if TagSelector(inbound.Tags).Matches(tags) {
			return true
		}
	}
	return false
}

type TagSelector map[string]string

func (s TagSelector) Matches(tags map[string]string) bool {
	for tag, value := range tags {
		inboundVal, exist := s[tag]
		if !exist || value != inboundVal {
			return false
		}
	}
	return true
}

func (d *Dataplane) Tags() map[string][]string {
	tags := map[string][]string{}
	for _, inbound := range d.Networking.Inbound {
		for tag, value := range inbound.Tags {
			exists := false
			for _, val := range tags[tag] {
				if val == value {
					exists = true
				}
			}
			if !exists {
				tags[tag] = append(tags[tag], value)
			}
		}
	}
	for _, values := range tags {
		sort.Strings(values)
	}
	return tags
}