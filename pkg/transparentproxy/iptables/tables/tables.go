package tables

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/chains"
	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type Table interface {
	Name() TableName
	Chains() []*chains.Chain
	CustomChains() []*chains.Chain
}

// BuildForRestore function generates all iptables rules for a given table in
// a format suitable for restoration using the `iptables-restore` command.
//
// For existing chains, it calls the `BuildForRestore(verbose)` method on each
// object to retrieve the individual rule restore commands.
//
// For custom chains, it first builds the command to create the chain using the
// appropriate flag based on the `verbose` flag (--new-chain or -N). Then, it
// calls `BuildForRestore(verbose)` on each new chain to retrieve the individual
// rule restore commands.
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
func BuildRulesForRestore(table Table, verbose bool) string {
	tableLine := fmt.Sprintf("* %s", table.Name())
	var customChainLines []string
	var ruleLines []string

	for _, c := range table.Chains() {
		ruleLines = append(ruleLines, c.BuildForRestore(verbose)...)
	}

	for _, c := range table.CustomChains() {
		customChainLines = append(customChainLines, fmt.Sprintf("%s %s", Flags[FlagNewChain][verbose], c.Name()))
		ruleLines = append(ruleLines, c.BuildForRestore(verbose)...)
	}

	if verbose {
		if len(customChainLines) > 0 {
			customChainLines = append(
				[]string{"# Custom Chains:"},
				customChainLines...,
			)
		}

		if len(ruleLines) > 0 {
			ruleLines = append([]string{"# Rules:"}, ruleLines...)
		}
	}

	lines := []string{tableLine}

	customChainsResult := strings.Join(customChainLines, "\n")
	if customChainsResult != "" {
		lines = append(lines, customChainsResult)
	}

	rulesResult := strings.Join(ruleLines, "\n")
	if rulesResult == "" {
		return ""
	}

	lines = append(lines, rulesResult, "COMMIT")

	if verbose {
		return strings.Join(lines, "\n\n")
	}

	return strings.Join(lines, "\n")
}
