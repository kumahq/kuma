package table

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chain"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type TableBuilder struct {
	name string

	newChains []*chain.Chain
	chains    []*chain.Chain
}

// Build
// TODO (bartsmykla): refactor
// TODO (bartsmykla): add tests
func (b *TableBuilder) Build(verbose bool) string {
	tableLine := fmt.Sprintf("* %s", b.name)
	var newChainLines []string
	var ruleLines []string

	for _, c := range b.chains {
		rules := c.Build(verbose)
		ruleLines = append(ruleLines, rules...)
	}

	for _, c := range b.newChains {
		newChainLines = append(newChainLines, fmt.Sprintf("%s %s", Flags["new-chain"][verbose], c.Name()))
		rules := c.Build(verbose)
		ruleLines = append(ruleLines, rules...)
	}

	if verbose {
		if len(newChainLines) > 0 {
			newChainLines = append(
				[]string{"# Custom Chains:"},
				newChainLines...,
			)
		}

		if len(ruleLines) > 0 {
			ruleLines = append([]string{"# Rules:"}, ruleLines...)
		}
	}

	lines := []string{tableLine}

	newChains := strings.Join(newChainLines, "\n")
	if newChains != "" {
		lines = append(lines, newChains)
	}

	rules := strings.Join(ruleLines, "\n")
	if rules == "" {
		return ""
	}

	lines = append(lines, rules, "COMMIT")

	if verbose {
		return strings.Join(lines, "\n\n")
	}

	return strings.Join(lines, "\n")
}
