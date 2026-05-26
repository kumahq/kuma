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
//
// Priors provably disjoint from self are dropped: if a prior cannot match when
// self matches, its negation is tautologically true and would only bloat the
// expression (and the CEL parser's recursion depth).
func ComposeExpr(self *common_api.Match, priors []*common_api.Match) string {
	selfExpr := MatchToCEL(self)
	var negatedPriors []string
	for _, p := range priors {
		if disjoint(self, p) {
			continue
		}
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

// disjoint reports whether two matches can never be simultaneously true. Sound
// but not complete: returns false when uncertain. A nil or empty sub-matcher
// on either side cannot prove disjointness on that dimension.
func disjoint(a, b *common_api.Match) bool {
	if a == nil || b == nil {
		return false
	}
	return spiffeDisjoint(a.SpiffeID, b.SpiffeID) || sniDisjoint(a.SNI, b.SNI)
}

func spiffeDisjoint(a, b *common_api.SpiffeIDMatch) bool {
	if a == nil || b == nil {
		return false
	}
	switch {
	case a.Type == common_api.ExactMatchType && b.Type == common_api.ExactMatchType:
		return a.Value != b.Value
	case a.Type == common_api.ExactMatchType && b.Type == common_api.PrefixMatchType:
		return !strings.HasPrefix(a.Value, b.Value)
	case a.Type == common_api.PrefixMatchType && b.Type == common_api.ExactMatchType:
		return !strings.HasPrefix(b.Value, a.Value)
	case a.Type == common_api.PrefixMatchType && b.Type == common_api.PrefixMatchType:
		return !strings.HasPrefix(a.Value, b.Value) && !strings.HasPrefix(b.Value, a.Value)
	}
	return false
}

func sniDisjoint(a, b *common_api.SNIMatch) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Value != b.Value
}

// celString returns a double-quoted CEL string literal with backslashes and
// double quotes escaped.
func celString(v string) string {
	escaped := strings.ReplaceAll(v, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	return `"` + escaped + `"`
}
