package firewalld

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// As specified in https://firewalld.org/documentation/man-pages/firewalld.direct.html

const (
	defaultFirewalldDirectPath = "/etc/firewalld/direct.xml"
	defaultRulenum             = 3
)

type IptablesRule struct {
	Mode          string
	Chain         string
	Rulenum       int
	Specification string
}

type IptablesTranslator struct {
	dryRun         bool
	output         io.Writer
	directFilePath string
	ruleParser     *regexp.Regexp
}

func (t *IptablesTranslator) WithDirectFilePath(filePath string) *IptablesTranslator {
	t.directFilePath = filePath

	return t
}

func (t *IptablesTranslator) WithDryRun(dryRun bool) *IptablesTranslator {
	t.dryRun = dryRun

	return t
}

func (t *IptablesTranslator) WithOutput(output io.Writer) *IptablesTranslator {
	t.output = output

	return t
}

func filterOutEmptyAndCommentLines(lines []string) []string {
	var result []string

	for _, line := range lines {
		if line != "" && !strings.HasPrefix(line, "#") {
			result = append(result, line)
		}
	}

	return result
}

func (t *IptablesTranslator) StoreRules(rawIptables string) (string, error) {
	direct, err := t.getPersistentDirect()
	if err != nil {
		return "", err
	}

	for table, rawRules := range parseIptablesRawInput(rawIptables) {
		for _, rawRule := range filterOutEmptyAndCommentLines(rawRules) {
			rule := t.translateRule(rawRule)

			switch rule.Mode {
			case "N", "new-chain":
				direct.AddChain(NewIP4Chain(table, rule.Chain))
			case "A", "append":
				direct.AddRule(NewIP4Rule(
					table,
					rule.Rulenum,
					rule.Chain,
					rule.Specification,
				))
			default:
				return "", fmt.Errorf("unsupported iptables mode [%s]", rule.Mode)
			}
		}
	}

	return t.store(direct)
}

func (t *IptablesTranslator) translateRule(rule string) IptablesRule {
	var result IptablesRule

	match := t.ruleParser.FindStringSubmatch(rule)

	for i, name := range t.ruleParser.SubexpNames() {
		if i != 0 && name != "" {
			switch name {
			case "mode":
				result.Mode = match[i]
			case "chain":
				result.Chain = match[i]
			case "rulenum":
				rulenum, err := strconv.Atoi(match[i])
				if err != nil || rulenum == 0 {
					rulenum = defaultRulenum
				}

				result.Rulenum = rulenum
			case "specification":
				result.Specification = match[i]
			}
		}
	}

	return result
}

func (t *IptablesTranslator) getPersistentDirect() (*Direct, error) {
	result := NewDirect()

	if t.dryRun {
		return result, nil
	}

	if _, err := os.Stat(t.directFilePath); err != nil {
		if os.IsPermission(err) {
			return nil, err
		}

		if os.IsNotExist(err) {
			// file does not exist, return an empty line
			return result, nil
		}
	}

	data, err := os.ReadFile(t.directFilePath)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (t *IptablesTranslator) store(direct *Direct) (string, error) {
	content := "\n\n" + direct.String() + "\n\n"

	if !t.dryRun {
		// -rw-r--r--.  1 root root  191 Mar 18 07:58 direct.xml
		if err := os.WriteFile(t.directFilePath, direct.Bytes(), 0o600); err != nil {
			return direct.String(), err
		}

		content += "iptables saved with firewalld" + "\n\n"
	}

	_, _ = t.output.Write([]byte(content))

	return direct.String(), nil
}

func parseIptablesRawInput(input string) map[string][]string {
	tableParser := regexp.MustCompile(`\* (?P<table>\w*)`)
	scanner := bufio.NewScanner(strings.NewReader(input))
	rules := map[string][]string{}
	table := ""

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "COMMIT") {
			table = ""

			continue
		}

		if matches := tableParser.FindStringSubmatch(line); len(matches) > 1 {
			table = matches[tableParser.SubexpIndex("table")]

			continue
		}

		// filter out empty and comment lines
		if table != "" && line != "" && !strings.HasPrefix(line, "#") {
			rules[table] = append(rules[table], line)
		}
	}

	return rules
}

func NewIptablesTranslator() *IptablesTranslator {
	return &IptablesTranslator{
		output:         io.Discard,
		directFilePath: defaultFirewalldDirectPath,
		ruleParser: regexp.MustCompile(
			`--?(?P<mode>[A-Za-z-]+)\s*(?P<chain>\w*)\s*(?P<rulenum>\d+)?\s*(?P<specification>.*)?`,
		),
	}
}
