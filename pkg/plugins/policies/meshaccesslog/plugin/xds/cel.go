package xds

import (
	"fmt"
	"strconv"
	"strings"

	common_api "github.com/kumahq/kuma/v2/api/common/v1alpha1"
)

// MatchToCEL converts a Kuma inbound Match into a CEL expression evaluating to
// true when the Envoy connection satisfies all sub-matchers in the Match.
// Returns "" for a nil/empty Match (which conceptually matches all traffic).
func MatchToCEL(m *common_api.Match) string {
	if m == nil {
		return ""
	}
	var parts []string
	if m.SpiffeID != nil {
		switch m.SpiffeID.Type {
		case common_api.ExactMatchType:
			parts = append(parts, fmt.Sprintf(`connection.uri_san_peer_certificate == %s`, strconv.Quote(m.SpiffeID.Value)))
		case common_api.PrefixMatchType:
			parts = append(parts, fmt.Sprintf(`connection.uri_san_peer_certificate.startsWith(%s)`, strconv.Quote(m.SpiffeID.Value)))
		}
	}
	if m.SNI != nil && m.SNI.Type == common_api.SNIExactMatchType {
		parts = append(parts, fmt.Sprintf(`connection.requested_server_name == %s`, strconv.Quote(m.SNI.Value)))
	}
	return strings.Join(parts, " && ")
}
