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

// BuildForRestore function generates all iptables rules for a given table in
// a format suitable for restoration using the `iptables-restore` command.
//
// For existing chains, it calls the `BuildForRestore(verbose)` method on each
// object to retrieve the individual rule restore commands.
//
// For newly created chains, it first builds the command to create the chain
// using the appropriate flag based on the `verbose` flag (--new-chain or -N).
// Then, it calls `BuildForRestore(verbose)` on each new chain to retrieve
// the individual rule restore commands.
//
// The built commands are then organized:
//   - A table line with the table name prefixed by "*".
//   - Optional section header "# Custom Chains:" with a newline (only for
//     verbose mode and if there are new chains).
//   - Individual commands to create new chains (only for new chains).
//   - Optional section header "# Rules:" with a newline (only for verbose mode
//     and if there are rules).
//   - Individual commands to restore rules from existing and new chains,
//     concatenated with newlines.
//   - "COMMIT" command.
//
// TODO (bartsmykla): refactor
// TODO (bartsmykla): add tests
func (b *TableBuilder) BuildForRestore(verbose bool) string {
	tableLine := fmt.Sprintf("* %s", b.name)
	var newChainLines []string
	var ruleLines []string

	for _, c := range b.chains {
		rules := c.BuildForRestore(verbose)
		ruleLines = append(ruleLines, rules...)
	}

	for _, c := range b.newChains {
		newChainLines = append(newChainLines, fmt.Sprintf("%s %s", Flags["new-chain"][verbose], c.Name()))
		rules := c.BuildForRestore(verbose)
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
