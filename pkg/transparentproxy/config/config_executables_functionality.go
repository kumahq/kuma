package config

import (
	"context"
	std_errors "errors"
	"strings"

	"github.com/pkg/errors"

	. "github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
)

type FunctionalityTables struct {
	Nat    bool
	Raw    bool
	Mangle bool
}

type FunctionalityModules struct {
	Tcp bool
	Udp bool
	// This module attempts to match various characteristics of the packet
	// creator, for locally-generated packets. It is only valid in the OUTPUT
	// chain, and even this some packets (such as ICMP ping responses) may have
	// no owner, and hence never match.
	// ref. iptables-extensions(8) > owner
	Owner bool
	// This module allows us to add comments (up to 256 characters) to any
	// iptables rule.
	// ref. iptables-extensions(8) > comment
	Comment bool
	// This module, when combined with connection tracking, allows access to
	// more connection tracking information than the "state" match.
	// ref. iptables-extensions(8) > conntrack
	Conntrack bool
	Multiport bool
}

type FunctionalityChains struct {
	DockerOutput bool
}

type FunctionalityRules struct {
	ExistingRules bool
}

type Functionality struct {
	Tables  FunctionalityTables
	Modules FunctionalityModules
	Chains  FunctionalityChains
	Rules   FunctionalityRules
}

// ConntrackZoneSplit checks if all required components are loaded to generate
// connection tracking zone split rules for DNS traffic.
//
// Requirements:
//   - 'nat' table: Used for generating DNS redirect rules (if necessary).
//   - 'raw' table: Needed to perform the actual traffic split.
//   - 'conntrack' module: Manages connection tracking for split traffic.
//   - 'udp' module: Processes UDP traffic (specifically DNS in this case).
//   - 'owner' module: Enables splitting traffic based on process ownership.
func (c Functionality) ConntrackZoneSplit() bool {
	return c.Tables.Nat && c.Tables.Raw && c.Modules.Conntrack &&
		c.Modules.Udp && c.Modules.Owner
}

// This function checks if the system meets the minimum requirements to install
// our transparent proxy. These requirements are:
// - Presence of a NAT table for manipulating network traffic.
// - Availability of two basic modules:
//   - owner: Allows matching packets based on the packet owner.
//   - tcp: Enables matching TCP packets.
func verifyMinimalRequirements(functionality Functionality) error {
	var errs []error

	if !functionality.Tables.Nat {
		errs = append(errs, errors.Errorf("missing table: %q", "nat"))
	}

	if !functionality.Modules.Tcp {
		errs = append(errs, errors.Errorf("missing module: %q", "tcp"))
	}

	if !functionality.Modules.Owner {
		errs = append(errs, errors.Errorf("missing module: %q", "owner"))
	}

	if len(errs) > 0 {
		return errors.Wrap(std_errors.Join(errs...), "unmet minimal requirements")
	}

	return nil
}

func verifyTable(
	ctx context.Context,
	iptablesSave InitializedExecutable,
	table TableName,
) (bool, string) {
	stdout, _, err := iptablesSave.Exec(ctx, FlagTable, string(table))
	if err != nil {
		return false, ""
	}

	return true, stdout.String()
}

func verifyModule(
	ctx context.Context,
	iptables InitializedExecutable,
	matchExtension string,
) bool {
	_, _, err := iptables.Exec(ctx, FlagMatch, matchExtension, FlagHelp)
	return err == nil
}

func verifyExistingRulesForTable(
	ctx context.Context,
	iptablesSave InitializedExecutable,
	table TableName,
	expectedLinesCount int,
) bool {
	stdout, _, err := iptablesSave.Exec(ctx, FlagTable, string(table))
	if err != nil {
		return false
	}

	var linesCount int
	for _, line := range strings.Split(stdout.String(), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "":
		case strings.HasPrefix(line, "#"): // comments
		case strings.HasPrefix(line, "*"): // table name (i.e. *nat)
		case strings.HasPrefix(line, "COMMIT"):
		default:
			linesCount++
		}
	}

	return linesCount > expectedLinesCount
}

func verifyExistingRules(ctx context.Context, iptablesSave InitializedExecutable) bool {
	// Iterate over each necessary iptables table and verify if the existing rules match the expected
	// number of lines for that table when only the default chains are present (i.e., no custom rules
	// have been added).

	// It is crucial to check the tables in a specific order. When iptables-save is run for a specified
	// table (e.g., iptables-save -t nat), it determines the order in which iptables-save will later
	// output the rules when called without specifying a table. For instance, if the following commands
	// are executed:
	//  ~ iptables-save -t nat
	//  ~ iptables-save -t mangle
	//  ~ iptables-save -t raw
	// any subsequent call to iptables-save (without specifying a table) will list the rules with tables
	// in the reverse order they were first checked (rules from the raw table will be shown first,
	// followed by the mangle table, and the nat table last).
	//
	// Because of this, we cannot iterate over a map like map[TableName]int, as it would produce
	// a random order in the output, potentially causing issues in tests that depend on a specific order,
	// such as the golden files in transparent proxy tests.
	//
	// The order below was chosen specifically to maintain consistency with existing golden files used in
	// transparent proxy tests. This order does not have any inherent semantic significance.

	// Check the nat table: :PREROUTING, :INPUT, :OUTPUT, :POSTROUTING
	if verifyExistingRulesForTable(ctx, iptablesSave, TableNat, 4) {
		return true
	}

	// Check the raw table: :PREROUTING, :OUTPUT
	if verifyExistingRulesForTable(ctx, iptablesSave, TableRaw, 2) {
		return true
	}

	// Check the mangle table: :PREROUTING, :INPUT, :FORWARD, :OUTPUT, :POSTROUTING
	if verifyExistingRulesForTable(ctx, iptablesSave, TableMangle, 5) {
		return true
	}

	return false
}

func verifyFunctionality(
	ctx context.Context,
	iptables InitializedExecutable,
	iptablesSave InitializedExecutable,
) (Functionality, error) {
	functionality := Functionality{
		Modules: FunctionalityModules{
			Owner:     verifyModule(ctx, iptables, ModuleOwner),
			Tcp:       verifyModule(ctx, iptables, ModuleTcp),
			Udp:       verifyModule(ctx, iptables, ModuleUdp),
			Comment:   verifyModule(ctx, iptables, ModuleComment),
			Conntrack: verifyModule(ctx, iptables, ModuleConntrack),
			Multiport: verifyModule(ctx, iptables, ModuleMultiport),
		},
		Rules: FunctionalityRules{
			ExistingRules: verifyExistingRules(ctx, iptablesSave),
		},
	}

	if ok, stdout := verifyTable(ctx, iptablesSave, TableNat); ok {
		functionality.Tables.Nat = true
		functionality.Chains.DockerOutput = DockerOutputChainRegex.MatchString(stdout)
	}

	functionality.Tables.Raw, _ = verifyTable(ctx, iptablesSave, TableRaw)
	functionality.Tables.Mangle, _ = verifyTable(ctx, iptablesSave, TableMangle)

	return functionality, verifyMinimalRequirements(functionality)
}
