package v1alpha1

import (
	"net"
	"strconv"
)

// Matches is simply an alias for MatchTags to make source code more aesthetic.
func (es *ExternalService) Matches(selector TagSelector) bool {
	if es != nil {
		return es.MatchTags(selector)
	}
	return false
}

func (es *ExternalService) MatchTags(selector TagSelector) bool {
	return selector.Matches(es.Tags)
}

func (es *ExternalService) MatchTagsFuzzy(selector TagSelector) bool {
	return selector.MatchesFuzzy(es.Tags)
}

func (es *ExternalService) GetService() string {
	if es == nil {
		return ""
	}
	return es.Tags[ServiceTag]
}

func (es *ExternalService) GetProtocol() string {
	if es == nil {
		return ""
	}
	return es.Tags[ProtocolTag]
}

func (es *ExternalService) GetHost() string {
	if es == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(es.Networking.Address)
	if err != nil {
		return ""
	}
	return host
}

func (es *ExternalService) GetPort() string {
	if es == nil {
		return ""
	}
	_, port, err := net.SplitHostPort(es.Networking.Address)
	if err != nil {
		return ""
	}
	return port
}

func (es *ExternalService) GetPortUInt32() uint32 {
	port := es.GetPort()
	iport, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(iport)
}

func (es *ExternalService) TagSet() SingleValueTagSet {
	return es.Tags
}
