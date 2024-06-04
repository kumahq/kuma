package v1alpha1

import (
	"text/template"

	"github.com/pkg/errors"

	"github.com/kumahq/kuma/pkg/core/validators"
)

func (r *HostnameGeneratorResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("template"), validateTemplate(r.Spec.Template))
	return verr.OrNil()
}

func validateTemplate(tmpl string) validators.ValidationError {
	var verr validators.ValidationError
	_, err := template.New("").Parse(tmpl)
	if err != nil {
		verr.AddViolationAt(validators.Root(), errors.Wrap(err, "couldn't parse template").Error())
	}
	return verr
}
