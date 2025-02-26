package parse

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/pkg/errors"

	core_model "github.com/kumahq/kuma/pkg/core/resources/model"
)

type ResourceScope string

const (
	Mesh   ResourceScope = "Mesh"
	Global ResourceScope = "Global"
)

type PolicyConfig struct {
	Package                      string
	Name                         string
	NameLower                    string
	Plural                       string
	SkipRegistration             bool
	SkipGetDefault               bool
	SingularDisplayName          string
	PluralDisplayName            string
	ShortName                    string
	Path                         string
	AlternativeNames             []string
	HasTo                        bool
	HasFrom                      bool
	HasRules                     bool
	HasStatus                    bool
	GoModule                     string
	ResourceDir                  string
	IsPolicy                     bool
	KDSFlags                     string
	Scope                        ResourceScope
	AllowedOnSystemNamespaceOnly bool
	IsReferenceableInTo          bool
	KubebuilderMarkers           []string
	InterpretFromEntriesAsRules  bool
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

	st, ok := mainStruct.Type.(*ast.StructType)
	if !ok {
		return PolicyConfig{}, errors.Errorf("type %s is not a struct", mainStruct.Name.String())
	}

	fields := map[string]bool{}
	for _, field := range st.Fields.List {
		for _, name := range field.Names {
			fields[name.Name] = true
		}
	}

	markers, kubebuilderMarkers, err := parseMarkers(mainComment)
	if err != nil {
		return PolicyConfig{}, err
	}

	cfg, err := newPolicyConfig(packageName, mainStruct.Name.String(), markers, fields)
	if err != nil {
		return PolicyConfig{}, err
	}
	cfg.KubebuilderMarkers = kubebuilderMarkers
	return cfg, nil
}

func parseMarkers(cg *ast.CommentGroup) (map[string]string, []string, error) {
	result := map[string]string{}
	var kubebuilderMarkers []string
	for _, comment := range cg.List {
		if !strings.HasPrefix(comment.Text, "// +") {
			continue
		}
		if strings.HasPrefix(comment.Text, "// +kubebuilder") {
			kubebuilderMarkers = append(kubebuilderMarkers, comment.Text)
			continue
		}
		trimmed := strings.TrimPrefix(comment.Text, "// +")
		mrkr := strings.Split(trimmed, "=")
		if len(mrkr) != 2 {
			return nil, nil, errors.Errorf("marker %s has wrong format", trimmed)
		}
		result[mrkr[0]] = mrkr[1]
	}
	return result, kubebuilderMarkers, nil
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

func newPolicyConfig(pkg, name string, markers map[string]string, fields map[string]bool) (PolicyConfig, error) {
	res := PolicyConfig{
		Package:             pkg,
		Name:                name,
		NameLower:           strings.ToLower(name),
		SingularDisplayName: core_model.DisplayName(name),
		PluralDisplayName:   core_model.PluralType(core_model.DisplayName(name)),
		HasTo:               fields["To"],
		HasFrom:             fields["From"],
		HasRules:            fields["Rules"],
		IsPolicy:            true,
	}

	if v, ok := parseBool(markers, "kuma:policy:skip_registration"); ok {
		res.SkipRegistration = v
	}
	if v, ok := parseBool(markers, "kuma:policy:skip_get_default"); ok {
		res.SkipGetDefault = v
	}
	if v, ok := parseBool(markers, "kuma:policy:is_policy"); ok {
		res.IsPolicy = v
	}
	if v, ok := parseBool(markers, "kuma:policy:has_status"); ok {
		res.HasStatus = v
	}
	if v, ok := parseBool(markers, "kuma:policy:allowed_on_system_namespace_only"); ok {
		res.AllowedOnSystemNamespaceOnly = v
	}
	if v, ok := parseBool(markers, "kuma:policy:is_referenceable_in_to"); ok {
		res.IsReferenceableInTo = v
	}
	if v, ok := parseBool(markers, "kuma:policy:interpret_from_entries_as_rules"); ok {
		res.InterpretFromEntriesAsRules = v
	}
	if v, ok := markers["kuma:policy:kds_flags"]; ok {
		res.KDSFlags = v
	} else if res.HasTo {
		// potentially a producer policy, so we need to sync it from one zone to another
		res.KDSFlags = "model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag | model.GlobalToAllButOriginalZoneFlag"
	} else {
		res.KDSFlags = "model.GlobalToAllZonesFlag | model.ZoneToGlobalFlag"
	}
	if v, ok := markers["kuma:policy:scope"]; ok {
		switch v {
		case "Global":
			res.Scope = Global
		case "Mesh":
			res.Scope = Mesh
		default:
			return res, errors.Errorf("couldn't parse %s as scope `Global` or `Mesh`", v)
		}
	} else {
		res.Scope = Mesh
	}

	if v, ok := markers["kuma:policy:singular_display_name"]; ok {
		res.SingularDisplayName = v
		res.PluralDisplayName = core_model.PluralType(v)
	}

	if v, ok := markers["kuma:policy:plural"]; ok {
		res.Plural = v
	} else {
		res.Plural = core_model.PluralType(res.Name)
	}

	if v, ok := markers["kuma:policy:short_name"]; ok {
		res.ShortName = v
	} else {
		var result []rune
		for _, char := range res.SingularDisplayName {
			if unicode.IsUpper(char) {
				result = append(result, unicode.ToLower(char))
			}
		}
		res.ShortName = string(result)
	}
	res.Path = strings.ToLower(res.Plural)

	return res, nil
}
