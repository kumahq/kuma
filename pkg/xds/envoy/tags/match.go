package tags

import (
	"fmt"
	"strings"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	core_policy "github.com/kumahq/kuma/pkg/core/policy"
)

func MatchingRegex(tags mesh_proto.SingleValueTagSet) (re string) {
	for _, key := range tags.Keys() {
		keyIsEqual := fmt.Sprintf(`&%s=`, key)
		var value string
		switch tags[key] {
		case "*":
			value = ``
		default:
			value = fmt.Sprintf(`[^&]*%s[,&]`, tags[key])
		}
		value = strings.ReplaceAll(value, ".", `\.`)
		expr := keyIsEqual + value + `.*`
		re += expr
	}
	re = `.*` + re
	return
}

func RegexOR(r ...string) string {
	if len(r) == 0 {
		return ""
	}
	if len(r) == 1 {
		return r[0]
	}
	return fmt.Sprintf("(%s)", strings.Join(r, "|"))
}

func MatchSourceRegex(policy core_policy.ConnectionPolicy) string {
	var selectorRegexs []string
	for _, selector := range policy.Sources() {
		selectorRegexs = append(selectorRegexs, MatchingRegex(selector.Match))
	}
	return RegexOR(selectorRegexs...)
}
