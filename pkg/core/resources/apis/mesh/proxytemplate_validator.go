package mesh

import (
	"fmt"
	"github.com/Kong/kuma/api/mesh/v1alpha1"
	"github.com/Kong/kuma/pkg/core/validators"
	"github.com/Kong/kuma/pkg/util/envoy"
	"github.com/Kong/kuma/pkg/xds/template"
	"strings"
)

var availableProfiles map[string]bool
var availableProfilesMsg string

func init() {
	profiles := []string{}
	availableProfiles = map[string]bool{}
	for _, profile := range template.AvailableProfiles {
		availableProfiles[profile] = true
		profiles = append(profiles, profile)
	}
	availableProfilesMsg = strings.Join(profiles, ",")
}

func (t *ProxyTemplateResource) Validate() error {
	var verr validators.ValidationError
	verr.AddError("", validateImports(t.Spec.Imports))
	verr.AddError("", validateResources(t.Spec.Resources))
	verr.AddError("", validateSelectors(t.Spec.Selectors))
	return verr.OrNil()
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
	var verr validators.ValidationError
	for i, selector := range selectors {
		if len(selector.Match) == 0 {
			verr.AddViolationAt(validators.RootedAt("selectors").Index(i), "has to contain at least one tag")
		}
		for key, value := range selector.Match {
			if key == "" {
				verr.AddViolationAt(validators.RootedAt("selectors").Index(i).Key(key), "tag cannot be empty")
			}
			if value == "" {
				verr.AddViolationAt(validators.RootedAt("selectors").Index(i).Key(key), "value of tag cannot be empty")
			}
		}
	}
	return verr
}
