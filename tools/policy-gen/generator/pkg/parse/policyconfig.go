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
	SkipDefault         bool
	SingularDisplayName string
	PluralDisplayName   string
	Path                string
	AlternativeNames    []string
	HasTo               bool
	HasFrom             bool
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

	hasTo := false
	hasFrom := false

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

	st, ok := mainStruct.Type.(*ast.StructType)
	if !ok {
		return PolicyConfig{}, errors.Errorf("type %s is not a struct", mainStruct.Name.String())
	}

	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			switch name.Name {
			case "From":
				hasFrom = true
			case "To":
				hasTo = true
			}
		}
	}

	markers, err := parseMarkers(mainComment)
	if err != nil {
		return PolicyConfig{}, err
	}

	return newPolicyConfig(packageName, mainStruct.Name.String(), markers, hasTo, hasFrom)
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

func parseBool(markers map[string]string, key string) (bool, bool) {
	if v, ok := markers[key]; ok {
		vbool, err := strconv.ParseBool(v)
		if err != nil {
			return false, false
		}
		return vbool, true
	}

	return false, false
}

func newPolicyConfig(pkg, name string, markers map[string]string, hasTo, hasFrom bool) (PolicyConfig, error) {
	res := PolicyConfig{
		Package:             pkg,
		Name:                name,
		NameLower:           strings.ToLower(name),
		SingularDisplayName: core_model.DisplayName(name),
		PluralDisplayName:   core_model.PluralDisplayName(name),
		AlternativeNames:    []string{strings.ToLower(name)},
		HasTo:               hasTo,
		HasFrom:             hasFrom,
	}

	if v, ok := parseBool(markers, "kuma:policy:skip_registration"); ok {
		res.SkipRegistration = v
	}
	if v, ok := parseBool(markers, "kuma:policy:skip_default"); ok {
		res.SkipDefault = v
	}

	if v, ok := markers["kuma:policy:plural"]; ok {
		res.Plural = v
	} else {
		res.Plural = core_model.PluralType(res.Name)
	}

	res.Path = strings.ToLower(res.Plural)

	return res, nil
}
