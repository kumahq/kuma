package v1alpha1

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	k8s_validation "k8s.io/apimachinery/pkg/util/validation"

	"github.com/kumahq/kuma/v3/pkg/core/validators"
)

func (r *HostnameGeneratorResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.Add(validateSelectors(path.Field("selector"), r.Spec.Selector))
	verr.AddErrorAt(path.Field("template"), validateTemplate(r.Spec.Template))
	verr.Add(validateExtension(path.Field("extension"), r.Spec.Extension))
	return verr.OrNil()
}

func validateSelectors(path validators.PathBuilder, selector Selector) validators.ValidationError {
	var verr validators.ValidationError
	selectorsDefined := 0
	if selector.MeshService != nil {
		selectorsDefined++
	}
	if selector.MeshExternalService != nil {
		selectorsDefined++
	}
	if selector.MeshMultiZoneService != nil {
		selectorsDefined++
	}
	if selectorsDefined != 1 {
		verr.AddViolationAt(path, "exact one selector (meshService, meshExternalService) must be defined")
	}
	return verr
}

func validateTemplate(tmpl string) validators.ValidationError {
	var verr validators.ValidationError
	if tmpl == "" {
		verr.AddViolationAt(validators.Root(), validators.MustNotBeEmpty)
		return verr
	}
	parsed, err := template.New("").
		Funcs(map[string]any{"label": func(string) (string, error) { return "label", nil }}).
		Parse(tmpl)
	if err != nil {
		verr.AddViolationAt(validators.Root(), errors.Wrap(err, "couldn't parse template").Error())
		return verr
	}
	stub := struct {
		Name        string
		DisplayName string
		Namespace   string
		Mesh        string
		Zone        string
	}{
		Name:        "name",
		DisplayName: "displayname",
		Namespace:   "namespace",
		Mesh:        "mesh",
		Zone:        "zone",
	}
	var sb strings.Builder
	if err := parsed.Execute(&sb, stub); err != nil {
		verr.AddViolationAt(validators.Root(), errors.Wrap(err, "couldn't render template with stub values").Error())
		return verr
	}
	rendered := sb.String()
	if violations := k8s_validation.IsDNS1123Subdomain(rendered); len(violations) > 0 {
		verr.AddViolationAt(validators.Root(), fmt.Sprintf("template renders to %q which is not a valid DNS name: %s", rendered, strings.Join(violations, ", ")))
	}
	return verr
}

func validateExtension(path validators.PathBuilder, extension *Extension) validators.ValidationError {
	var verr validators.ValidationError
	if extension != nil && extension.Type == "" {
		verr.AddViolationAt(path.Field("type"), validators.MustNotBeEmpty)
	}
	return verr
}
