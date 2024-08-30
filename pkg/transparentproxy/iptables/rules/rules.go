package rules

import (
	"fmt"
	"strings"

	"github.com/kumahq/kuma/pkg/transparentproxy/config"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/consts"
	"github.com/kumahq/kuma/pkg/transparentproxy/iptables/parameters"
)

type Rule struct {
	table      consts.TableName
	chain      string
	position   uint
	parameters parameters.Parameters
	comment    string
}

type RuleBuilder struct {
	parameters parameters.Parameters
	comment    string
	insert     bool
}

func (b *RuleBuilder) WithComment(comment string) *RuleBuilder {
	b.comment = comment
	return b
}

func (b *RuleBuilder) WithCommentf(format string, a ...any) *RuleBuilder {
	b.comment = fmt.Sprintf(format, a...)
	return b
}

func (b *RuleBuilder) WithConditionalComment(
	condition bool,
	commentTrue string,
	commentFalse string,
) *RuleBuilder {
	if !condition {
		b.comment = commentFalse
	} else {
		b.comment = commentTrue
	}

	return b
}

func (b *RuleBuilder) Build(
	table consts.TableName,
	chain string,
	position uint,
) (*Rule, uint) {
	rule := &Rule{
		table:      table,
		chain:      chain,
		parameters: b.parameters,
		comment:    b.comment,
	}

	if b.insert {
		position++
		rule.position = position
	}

	return rule, position
}

// NewAppendRule creates a new RuleBuilder for an iptables rule that will be
// appended to the end of an existing chain. This function takes a variable
// number of parameters, each represented as a pointer to a Parameter object,
// which specify the various conditions and actions for the rule.
//
// Args:
//   - parameters (...*parameters.Parameter): A variadic list of pointers to
//     Parameter objects that define the rule's conditions and actions.
//
// Returns:
//   - *RuleBuilder: A pointer to a RuleBuilder configured to append a new rule
//     with the specified parameters.
func NewAppendRule(parameters ...*parameters.Parameter) *RuleBuilder {
	return &RuleBuilder{parameters: parameters}
}

// NewInsertRule creates a new RuleBuilder for an iptables rule that will be
// inserted at a specific position within an existing chain. This function takes
// a variable number of parameters, each represented as a pointer to a Parameter
// object, which specify the various conditions and actions for the rule.
// The rule will be marked for insertion rather than appending.
//
// Args:
//   - parameters (...*parameters.Parameter): A variadic list of pointers to
//     Parameter objects that define the rule's conditions and actions.
//
// Returns:
//   - *RuleBuilder: A pointer to a RuleBuilder configured to insert a new rule
//     with the specified parameters.
func NewInsertRule(parameters ...*parameters.Parameter) *RuleBuilder {
	return &RuleBuilder{parameters: parameters, insert: true}
}

// NewConditionalInsertOrAppendRule creates a new RuleBuilder for an iptables rule
// that will either be appended to the end of an existing chain or inserted at a
// specific position based on the given `insert` argument
func NewConditionalInsertOrAppendRule(insert bool, parameters ...*parameters.Parameter) *RuleBuilder {
	return &RuleBuilder{parameters: parameters, insert: insert}
}

// BuildForRestore generates an iptables rule formatted for use with
// `iptables-restore`.
//
// This function constructs the rule string based on the provided configuration
// settings (`cfg`). The `cfg.Verbose` flag determines whether long or short
// option names are used in the command (e.g., "--append" vs. "-A"). The rule's
// position within the chain is controlled by the `r.position` attribute:
//
// - If `r.position` is 0, the rule is appended to the chain.
//   - Uses "--append" or "-A" depending on `cfg.Verbose`.
//
// - If `r.position` is not 0, the rule is inserted at the specified position.
//   - Uses "--insert" or "-I" depending on `cfg.Verbose`.
//
// The command string is built as a space-separated list of components:
// the flag, the chain name, the optional position (if not zero), and the rule
// parameters. If a comment is provided (`r.comment`) and the comment
// functionality is enabled in the configuration, it is included in the
// parameters with the prefix specified by `consts.IptablesRuleCommentPrefix`.
//
// The `parameters.Build(cfg.Verbose)` method is used to generate the final
// parameter list in the appropriate format.
//
// Args:
//   - cfg (config.InitializedConfigIPvX): Configuration settings that control
//     the verbose output and other behaviors.
//
// Returns:
//   - string: The constructed iptables rule formatted for `iptables-restore`.
func (r *Rule) BuildForRestore(cfg config.InitializedConfigIPvX) string {
	flag := consts.FlagVariationsMap[consts.FlagAppend][cfg.Verbose]
	if r.position != 0 {
		flag = consts.FlagVariationsMap[consts.FlagInsert][cfg.Verbose]
	}

	cmd := []string{flag, r.chain}

	if r.position != 0 {
		cmd = append(cmd, fmt.Sprintf("%d", r.position))
	}

	var params parameters.Parameters

	if cfg.Comments.Enabled && r.comment != "" {
		params = append(
			params,
			parameters.Match(
				parameters.Comment(cfg.Comments.Prefix, r.comment),
			),
		)
	}

	params = append(params, r.parameters...)

	return strings.Join(append(cmd, params.Build(cfg.Verbose)...), " ")
}
