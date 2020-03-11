package mesh

import (
	"fmt"
	"strings"

	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/util/envoy"
)

var availableProfiles map[string]bool
var availableProfilesMsg string

func init() {
	profiles := []string{}
	availableProfiles = map[string]bool{}
	for _, profile := range AvailableProfiles {
		availableProfiles[profile] = true
		profiles = append(profiles, profile)
	}
	availableProfilesMsg = strings.Join(profiles, ",")
}

func (t *ProxyTemplateResource) Validate() error {
	var verr validators.ValidationError
	verr.Add(validateSelectors(t.Spec.Selectors))
	verr.AddError("conf", validateConfig(t.Spec.Conf))
	return verr.OrNil()
}

func validateConfig(conf *v1alpha1.ProxyTemplate_Conf) validators.ValidationError {
	var verr validators.ValidationError
	verr.Add(validateImports(conf.GetImports()))
	verr.Add(validateResources(conf.GetResources()))
	return verr
}

func validateImports(imports []string) validators.ValidationError {
	var verr validators.ValidationError
	for i, imp := range imports {
		if imp == "" {
			verr.AddViolationAt(validators.RootedAt("imports").Index(i), "cannot be empty")
			continue
		}
		if !availableProfiles[imp] {
			verr.AddViolationAt(validators.RootedAt("imports").Index(i), fmt.Sprintf("profile not found. Available profiles: %s", availableProfilesMsg))
		}
	}
	return verr
}

func validateResources(resources []*v1alpha1.ProxyTemplateRawResource) validators.ValidationError {
	var verr validators.ValidationError
	for i, resource := range resources {
		if resource.Name == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("name"), "cannot be empty")
		}
		if resource.Version == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("version"), "cannot be empty")
		}
		if resource.Resource == "" {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("resource"), "cannot be empty")
		} else if _, err := envoy.ResourceFromYaml(resource.Resource); err != nil {
			verr.AddViolationAt(validators.RootedAt("resources").Index(i).Field("resource"), fmt.Sprintf("native Envoy resource is not valid: %s", err.Error()))
		}
	}
	return verr
}

func validateSelectors(selectors []*v1alpha1.Selector) validators.ValidationError {
	return ValidateSelectors(validators.RootedAt("selectors"), selectors, ValidateSelectorsOpts{
		ValidateSelectorOpts: ValidateSelectorOpts{
			RequireService:       true,
			RequireAtLeastOneTag: true,
		},
	})
}
