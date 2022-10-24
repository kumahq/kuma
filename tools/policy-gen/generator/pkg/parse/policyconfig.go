package parse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type PolicyConfig struct {
	Package             string
	Name                string
	NameLower           string
	Plural              string
	SkipRegistration    bool
	SingularDisplayName string
	PluralDisplayName   string
	Path                string
	AlternativeNames    []string
}

func Policy(path string) (PolicyConfig, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return PolicyConfig{}, err
	}

	policyName := strings.Split(filepath.Base(path), ".")[0]
	var mainStruct *ast.TypeSpec
	var mainComment *ast.CommentGroup
	var packageName string

	ast.Inspect(f, func(n ast.Node) bool {
		if file, ok := n.(*ast.File); ok {
			packageName = file.Name.String()
			return true
		}
		if gd, ok := n.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			for _, spec := range gd.Specs {
				if strings.ToLower(spec.(*ast.TypeSpec).Name.String()) == policyName {
					mainStruct = spec.(*ast.TypeSpec)
					mainComment = gd.Doc
					return false
				}
			}
			return false
		}
		return false
	})

	markers, err := parseMarkers(mainComment)
	if err != nil {
		return PolicyConfig{}, err
	}

	return newPolicyConfig(packageName, mainStruct.Name.String(), markers)
}

func parseMarkers(cg *ast.CommentGroup) (map[string]string, error) {
	result := map[string]string{}
	for _, comment := range cg.List {
		if !strings.HasPrefix(comment.Text, "// +") {
			continue
		}
		trimmed := strings.TrimPrefix(comment.Text, "// +")
		mrkr := strings.Split(trimmed, "=")
		if len(mrkr) != 2 {
			return nil, errors.Errorf("marker %s has wrong format", trimmed)
		}
		result[mrkr[0]] = mrkr[1]
	}
	return result, nil
}

func newPolicyConfig(pkg, name string, markers map[string]string) (PolicyConfig, error) {
	res := PolicyConfig{
		Package:   pkg,
		Name:      name,
		NameLower: strings.ToLower(name),
	}

	if v, ok := markers["kuma:skip_registration"]; ok {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return PolicyConfig{}, err
		}
		res.SkipRegistration = vbool
	}

	if v, ok := markers["kuma:plural"]; ok {
		res.Plural = v
	} else {
		res.Plural = core_model.PluralType(res.Name)
	}

	res.SingularDisplayName = core_model.DisplayName(res.Name)
	res.PluralDisplayName = core_model.PluralDisplayName(res.Name)
	res.Path = strings.ToLower(res.Plural)
	res.AlternativeNames = []string{strings.ToLower(res.Name)}

	return res, nil
}
