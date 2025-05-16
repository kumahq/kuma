package v1alpha1

import (
	"text/template"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/validators"
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
	}
	_, err := template.New("").
		Funcs(map[string]any{"label": func(key string) (string, error) { return "", nil }}).
		Parse(tmpl)
	if err != nil {
		verr.AddViolationAt(validators.Root(), errors.Wrap(err, "couldn't parse template").Error())
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
