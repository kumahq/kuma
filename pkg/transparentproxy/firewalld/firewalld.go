package firewalld

import (
	"encoding/xml"
	"os"
	"regexp"
	"strconv"

	"github.com/pkg/errors"
)

// As specified in https://firewalld.org/documentation/man-pages/firewalld.direct.html

const defaultFirewalldDirectPath = "/etc/firewalld/direct.xml"

type FirewalldIptablesTranslator struct {
	dryRun         bool
	directlXMLPath string
	parser         *regexp.Regexp
}

const (
	iptablesMode          = "mode"
	iptablesChain         = "chain"
	iptablesRulenum       = "rulenum"
	iptablesSpecification = "specification"
	defaultRulenum        = 3
)

func (fit *FirewalldIptablesTranslator) StoreRules(rules map[string][]string) (string, error) {
	direct, err := fit.getPersistentDirect()
	if err != nil {
		return "", err
	}

	for table, rules := range rules {
		for _, rule := range rules {
			translated, err := fit.translateRule(rule)
			if err != nil {
				return "", err
			}

			mode := translated[iptablesMode]
			chain := translated[iptablesChain]
			rulenum, err := strconv.Atoi(translated[iptablesRulenum])
			if err != nil || rulenum == 0 {
				rulenum = defaultRulenum
			}
			specification := translated[iptablesSpecification]

			switch mode {
			case "N":
				direct.AddChain(NewIP4Chain(table, chain))
			case "A":
				direct.AddRule(NewIP4Rule(rulenum, table, chain, specification))
			default:
				return "", errors.Errorf("unuspported iptable mode [%s]", mode)
			}
		}
	}

	return fit.store(direct)
}

func (fit *FirewalldIptablesTranslator) translateRule(rule string) (map[string]string, error) {
	match := fit.parser.FindStringSubmatch(rule)
	result := make(map[string]string)
	for i, name := range fit.parser.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	return result, nil
}

func (fit *FirewalldIptablesTranslator) getPersistentDirect() (*Direct, error) {
	result := NewDirect()

	if fit.dryRun {
		return result, nil
	}

	_, err := os.Stat(fit.directlXMLPath)
	if err != nil {
		if os.IsPermission(err) {
			return nil, err
		}
		if os.IsNotExist(err) {
			// file does not exist, return an empty line
			return result, nil
		}
	}

	data, err := os.ReadFile(fit.directlXMLPath)
	if err != nil {
		return nil, err
	}

	err = xml.Unmarshal(data, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (fit *FirewalldIptablesTranslator) store(direct *Direct) (string, error) {
	if !fit.dryRun {
		err := os.WriteFile(fit.directlXMLPath, direct.Bytes(), 0644) // -rw-r--r--.  1 root root  191 Mar 18 07:58 direct.xml
		if err != nil {
			return "", err
		}
	}

	return direct.String(), nil
}

func NewFirewalldIptablesTranslator(dryRun bool) *FirewalldIptablesTranslator {
	return &FirewalldIptablesTranslator{
		dryRun:         dryRun,
		directlXMLPath: defaultFirewalldDirectPath,
		parser:         regexp.MustCompile(`-(?P<mode>[A-Z])\s*(?P<chain>\w*)\s*(?P<rulenum>\d+)?\s*(?P<specification>.*)?`),
	}
}
