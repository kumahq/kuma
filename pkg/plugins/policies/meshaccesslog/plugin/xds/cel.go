package xds

import (
	"fmt"
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
			parts = append(parts, fmt.Sprintf(`connection.uri_san_peer_certificate == %s`, celString(m.SpiffeID.Value)))
		case common_api.PrefixMatchType:
			parts = append(parts, fmt.Sprintf(`connection.uri_san_peer_certificate.startsWith(%s)`, celString(m.SpiffeID.Value)))
		}
	}
	if m.SNI != nil && m.SNI.Type == common_api.SNIExactMatchType {
		parts = append(parts, fmt.Sprintf(`connection.requested_server_name == %s`, celString(m.SNI.Value)))
	}
	return strings.Join(parts, " && ")
}

// ComposeExpr returns the CEL expression for a "first-match-wins" rule:
// the rule applies only when its own match holds AND none of the more-specific
// prior rules' matches hold. Returns "" if no filter is needed (no self match
// and no priors — i.e. the only rule with no constraints).
func ComposeExpr(self *common_api.Match, priors []*common_api.Match) string {
	selfExpr := MatchToCEL(self)
	var negatedPriors []string
	for _, p := range priors {
		priorExpr := MatchToCEL(p)
		if priorExpr == "" {
			// A prior rule that matches everything would shadow this rule entirely;
			// the sort places catch-all rules last, so this is not expected. Skip
			// defensively to keep the expression valid.
			continue
		}
		negatedPriors = append(negatedPriors, fmt.Sprintf("!(%s)", priorExpr))
	}
	switch {
	case selfExpr == "" && len(negatedPriors) == 0:
		return ""
	case selfExpr == "":
		return strings.Join(negatedPriors, " && ")
	case len(negatedPriors) == 0:
		return selfExpr
	default:
		return selfExpr + " && " + strings.Join(negatedPriors, " && ")
	}
}

// celString returns a double-quoted CEL string literal with backslashes and
// double quotes escaped.
func celString(v string) string {
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
