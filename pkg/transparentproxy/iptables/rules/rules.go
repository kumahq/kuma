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
	position   uint
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

func (b *RuleBuilder) WithPosition(position uint) *RuleBuilder {
	b.position = position
	return b
}

func (b *RuleBuilder) Build(table consts.TableName, chain string) *Rule {
	return &Rule{
		table:      table,
		chain:      chain,
		position:   b.position,
		parameters: b.parameters,
		comment:    b.comment,
	}
}

func NewRule(parameters ...*parameters.Parameter) *RuleBuilder {
	return &RuleBuilder{parameters: parameters}
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

	if cfg.Comment.Enabled && r.comment != "" {
		params = append(
			params,
			parameters.Match(
				parameters.Comment(cfg.Comment.Prefix, r.comment),
			),
		)
	}

	params = append(params, r.parameters...)

	return strings.Join(append(cmd, params.Build(cfg.Verbose)...), " ")
}
