package v1alpha1

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"

	core_model "github.com/kumahq/kuma/v3/pkg/core/resources/model"
	"github.com/kumahq/kuma/v3/pkg/core/validators"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

func (r *MeshIdentityResource) validate() error {
	var verr validators.ValidationError
	path := validators.RootedAt("spec")
	verr.AddErrorAt(path.Field("spiffeID"), validateSPIFFEID(pointer.Deref(r.Spec.SpiffeID)))
	if r.Spec.Provider != nil {
		verr.AddErrorAt(path.Field("provider"), validateProvider(pointer.Deref(r.Spec.Provider)))
	}
	return verr.OrNil()
}

// ValidateUpdate rejects edits to the SPIFFE ID. A trust domain is an identity
// namespace and its path names workloads inside it, so moving either is a
// migration, not an edit: issuance re-renders immediately while the MeshTrust
// publishing the CA bundle, MeshService.spec.identities and the SPIFFE ID rules
// of MeshTrafficPermission all move on their own schedules, leaving already
// issued leaves unverifiable in between. Users migrate by creating a second
// MeshIdentity under a different name and deleting the old one, which keeps both
// trust domains published for the whole transition. Reusing the name instead
// rebuilds the same gap, because the MeshTrust and the CA are keyed by it.
func (r *MeshIdentityResource) ValidateUpdate(previous core_model.Resource) error {
	prev, ok := previous.(*MeshIdentityResource)
	if !ok {
		return fmt.Errorf("invalid type %T for the previous MeshIdentity", previous)
	}
	var verr validators.ValidationError
	path := validators.RootedAt("spec").Field("spiffeID")
	previousSpiffeID := pointer.Deref(prev.Spec.SpiffeID)
	currentSpiffeID := pointer.Deref(r.Spec.SpiffeID)
	if previousTrustDomain, currentTrustDomain := pointer.Deref(previousSpiffeID.TrustDomain), pointer.Deref(currentSpiffeID.TrustDomain); previousTrustDomain != currentTrustDomain {
		verr.AddViolationAt(path.Field("trustDomain"), immutableFieldMessage(previousTrustDomain, currentTrustDomain))
	}
	if previousPath, currentPath := pointer.Deref(previousSpiffeID.Path), pointer.Deref(currentSpiffeID.Path); previousPath != currentPath {
		verr.AddViolationAt(path.Field("path"), immutableFieldMessage(previousPath, currentPath))
	}
	return verr.OrNil()
}

func immutableFieldMessage(previous string, current string) string {
	return fmt.Sprintf(
		"is immutable, cannot be changed from %q to %q. Create a MeshIdentity under a different name with the new value and delete this one once every workload has migrated",
		previous,
		current,
	)
}

// validateSPIFFEID renders each template with sample values so a template that
// the control plane cannot execute, or that renders into a SPIFFE ID the
// spiffe library rejects, is rejected here instead of failing identity
// generation for every dataplane the identity matches. The stubs carry a value
// for every field a template can reference: whether a dataplane actually
// provides one, such as the namespace on Universal, depends on where it runs,
// so a template that only renders with some fields empty stays valid.
func validateSPIFFEID(spiffeID SpiffeID) validators.ValidationError {
	var verr validators.ValidationError
	if trustDomain := pointer.Deref(spiffeID.TrustDomain); trustDomain != "" {
		rendered, err := renderTemplateStub(trustDomain, trustDomainTemplateData{
			Mesh: "mesh",
			Zone: "zone",
		})
		if err != nil {
			verr.AddViolation("trustDomain", err.Error())
		} else if _, err := spiffeid.TrustDomainFromString(rendered); err != nil {
			verr.AddViolation("trustDomain", fmt.Sprintf("template renders to %q which is not a valid SPIFFE ID trust domain: %s", rendered, err))
		}
	}
	if path := pointer.Deref(spiffeID.Path); path != "" {
		rendered, err := renderTemplateStub(path, spiffeIDTemplateData{
			TrustDomain:    sampleTrustDomain,
			Namespace:      "namespace",
			ServiceAccount: "service-account",
			Workload:       "workload",
		})
		if err != nil {
			verr.AddViolation("path", err.Error())
		} else if err := spiffeid.ValidatePath(rendered); err != nil {
			verr.AddViolation("path", fmt.Sprintf("template renders to %q which is not a valid SPIFFE ID path: %s", rendered, err))
		}
	}
	return verr
}

const sampleTrustDomain = "trust-domain"

// renderTemplateStub parses and executes a template the way issuance does. The
// label function resolves to a sample value because a template can reference
// any dataplane label and validation cannot know which ones a dataplane
// carries.
func renderTemplateStub(tmpl string, data any) (string, error) {
	parsed, err := template.New("").Funcs(map[string]any{
		"label": func(string) (string, error) { return "label", nil },
	}).Parse(tmpl)
	if err != nil {
		return "", errors.Wrap(err, "couldn't parse template")
	}
	var sb strings.Builder
	if err := parsed.Execute(&sb, data); err != nil {
		return "", errors.Wrap(err, "couldn't render template with stub values")
	}
	return sb.String(), nil
}

func validateProvider(provider Provider) validators.ValidationError {
	var verr validators.ValidationError
	switch provider.Type {
	case BundledType:
		verr.Add(validateBundled(validators.RootedAt("bundled"), provider.Bundled))
	case SpireType:
		verr.Add(validateSpire(validators.RootedAt("spire"), provider.Spire))
	case ExtensionType:
		verr.Add(validateExtension(validators.RootedAt("extension"), provider.Extension))
	default:
		verr.AddError("type", validators.MakeFieldMustBeOneOfErr(string(provider.Type), string(BundledType), string(SpireType), string(ExtensionType)))
	}
	return verr
}

func validateBundled(path validators.PathBuilder, b *Bundled) validators.ValidationError {
	var verr validators.ValidationError
	if b == nil {
		verr.AddViolationAt(path, "configuration needs to be defined")
		return verr
	}
	if b.MeshTrustCreation != nil {
		switch *b.MeshTrustCreation {
		case MeshTrustCreationEnabled:
		case MeshTrustCreationDisabled:
		default:
			verr.AddErrorAt(
				path.Field("meshTrustCreation"),
				validators.MakeFieldMustBeOneOfErr(string(*b.MeshTrustCreation), string(MeshTrustCreationEnabled), string(MeshTrustCreationDisabled)),
			)
		}
	}
	if b.Autogenerate != nil && pointer.Deref(b.Autogenerate.Enabled) {
		if b.CA != nil {
			verr.AddViolationAt(path.Field("ca"), "shouldn't be defined once using autogenerated")
		}
	} else {
		if b.CA == nil {
			verr.AddViolationAt(path.Field("ca"), validators.MustBeDefined)
		} else {
			if b.CA.Certificate == nil {
				verr.AddViolationAt(path.Field("ca").Field("certificate"), validators.MustBeDefined)
			} else {
				verr.Add(b.CA.Certificate.ValidateSecureDataSource(path.Field("ca").Field("certificate")))
			}
			if b.CA.PrivateKey == nil {
				verr.AddViolationAt(path.Field("ca").Field("privateKey"), validators.MustBeDefined)
			} else {
				verr.Add(b.CA.PrivateKey.ValidateSecureDataSource(path.Field("ca").Field("privateKey")))
			}
		}
	}
	if b.CertificateParameters != nil {
		if b.CertificateParameters.Expiry != nil {
			verr.Add(validators.ValidateDurationNotNegative(path.Field("certificateParameters").Field("expiry"), b.CertificateParameters.Expiry))
		}
	}
	return verr
}

func validateSpire(path validators.PathBuilder, b *Spire) validators.ValidationError {
	var verr validators.ValidationError
	if b == nil {
		verr.AddViolationAt(path, "configuration needs to be defined")
		return verr
	}
	if b.Agent != nil {
		verr.AddErrorAt(path.Field("agent"), validateSpireAgent(path.Field("agent"), b.Agent))
	}
	return verr
}

// validateExtension performs base validation for the Extension provider type.
// Extension-specific config validation (e.g. parsing config fields) is delegated to the
// extension implementation via registry.RegisterTypeValidator.
func validateExtension(path validators.PathBuilder, b *Extension) validators.ValidationError {
	var verr validators.ValidationError
	if b == nil {
		verr.AddViolationAt(path, "configuration needs to be defined")
		return verr
	}
	if b.Name == "" {
		verr.AddViolationAt(path.Field("name"), "extension name needs to be defined")
	}
	return verr
}

func validateSpireAgent(path validators.PathBuilder, b *SpireAgent) validators.ValidationError {
	return validators.ValidateDurationNotNegativeOrNil(path.Field("timeout"), b.Timeout)
}
