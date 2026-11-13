package v1alpha1

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
)

const (
	defaultTrustDomainTemplate           = "{{ .Mesh }}.{{ .Zone }}.mesh.local"
	defaultK8sSpiffeIDPathTemplate       = "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
	defaultUniversalSpiffeIDPathTemplate = "/workload/{{ .Workload }}"
)

// AllMatched returns a list of MeshIdentity policies that match the given labels and are initialized or in SpiffeIDProviderMode.
func AllMatched(
	labels map[string]string,
	meshIdentities []*MeshIdentityResource,
) []*MeshIdentityResource {
	var matches []*MeshIdentityResource
	for _, mi := range meshIdentities {
		if mi.Spec.Selector == nil || mi.Spec.Selector.Dataplane == nil || !mi.Spec.Selector.Dataplane.Matches(labels) {
			continue
		}
		matches = append(matches, mi)
	}

	return matches
}

// BestMatched returns the most specific MeshIdentity policy that matches the given labels.
// If multiple policies have the same specificity, the one with the greatest number of matching labels is selected first,
// and if still tied, the policy with the lexicographically smallest name is chosen.
func BestMatched(
	labels map[string]string,
	meshIdentities []*MeshIdentityResource,
) (*MeshIdentityResource, bool) {
	type scoredMatch struct {
		matchCount int
		name       string
		policy     *MeshIdentityResource
	}

	var matches []scoredMatch

	for _, mi := range AllMatched(labels, meshIdentities) {
		matchCount := len(pointer.Deref(mi.Spec.Selector.Dataplane.MatchLabels))
		matches = append(matches, scoredMatch{
			matchCount: matchCount,
			name:       mi.GetMeta().GetName(),
			policy:     mi,
		})
	}

	if len(matches) == 0 {
		return nil, false
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].matchCount != matches[j].matchCount {
			return matches[i].matchCount > matches[j].matchCount
		}
		return matches[i].name < matches[j].name
	})

	return matches[0].policy, true
}

// SpiffeIDPathTemplate returns the path portion of the SPIFFE ID template that
// will be used when rendering the SPIFFE ID for a dataplane. It falls back to
// an environment-specific default when the MeshIdentity does not specify a
// custom path.
func (i *MeshIdentity) SpiffeIDPathTemplate(env config_core.EnvironmentType) string {
	var defaultSpiffeIDPathTemplate string
	switch env {
	case config_core.KubernetesEnvironment:
		defaultSpiffeIDPathTemplate = defaultK8sSpiffeIDPathTemplate
	case config_core.UniversalEnvironment:
		defaultSpiffeIDPathTemplate = defaultUniversalSpiffeIDPathTemplate
	}
	if i.SpiffeID == nil {
		return defaultSpiffeIDPathTemplate
	}
	return pointer.DerefOr(i.SpiffeID.Path, defaultSpiffeIDPathTemplate)
}

func (i *MeshIdentity) getSpiffeIDTemplate(env config_core.EnvironmentType) string {
	builder := strings.Builder{}
	builder.WriteString("spiffe://")
	builder.WriteString("{{ .TrustDomain }}")
	builder.WriteString(i.SpiffeIDPathTemplate(env))
	return builder.String()
}

func (i *MeshIdentity) GetTrustDomain(meta model.ResourceMeta, localZone string) (string, error) {
	var trustDomainTmpl string
	if i.SpiffeID == nil || i.SpiffeID.TrustDomain == nil {
		trustDomainTmpl = defaultTrustDomainTemplate
	} else {
		trustDomainTmpl = pointer.Deref(i.SpiffeID.TrustDomain)
	}

	zone := meta.GetLabels()[mesh_proto.ZoneTag]
	if zone == "" {
		zone = localZone
	}

	data := struct {
		Mesh string
		Zone string
	}{
		Mesh: meta.GetMesh(),
		Zone: zone,
	}

	return renderTemplate(trustDomainTmpl, func(key string) (string, error) {
		return lookupLabel(meta.GetLabels(), key)
	}, data)
}

// spiffeIDInputs feeds the SPIFFE ID template. Inputs are methods, not fields,
// so a render records every label it reads, however the template spells it.
type spiffeIDInputs struct {
	trustDomain string
	labels      map[string]string
	read        map[string]struct{}
}

// Format keeps print, printf, js and the like from dumping the fields, which
// would expose labels without recording the read. Templates print a pointer's
// struct value, so this needs a value receiver.
func (spiffeIDInputs) Format(f fmt.State, _ rune) {
	_, _ = fmt.Fprint(f, "spiffeIDInputs")
}

func newSpiffeIDInputs(trustDomain string, meta model.ResourceMeta) *spiffeIDInputs {
	return &spiffeIDInputs{
		trustDomain: trustDomain,
		labels:      meta.GetLabels(),
		read:        map[string]struct{}{},
	}
}

func (in *spiffeIDInputs) TrustDomain() string { return in.trustDomain }

func (in *spiffeIDInputs) Namespace() string { return in.labelOrEmpty(mesh_proto.KubeNamespaceTag) }

func (in *spiffeIDInputs) ServiceAccount() string {
	return in.labelOrEmpty(metadata.KumaServiceAccount)
}

func (in *spiffeIDInputs) Workload() string { return in.labelOrEmpty(metadata.KumaWorkload) }

func (in *spiffeIDInputs) labelOrEmpty(key string) string {
	in.read[key] = struct{}{}
	return in.labels[key]
}

func (in *spiffeIDInputs) label(key string) (string, error) {
	in.read[key] = struct{}{}
	return lookupLabel(in.labels, key)
}

// RenderSpiffeID renders the SPIFFE ID this identity issues to the resource
// described by meta, in the trust domain it issues in. It also returns, sorted,
// the labels the render read, even ones meta lacks and even when it fails.
// Checks of those labels must use this so they see what issuance sees.
func (r *MeshIdentityResource) RenderSpiffeID(meta model.ResourceMeta, localZone string, environment config_core.EnvironmentType) (string, []string, error) {
	trustDomain, err := r.Spec.GetTrustDomain(r.GetMeta(), localZone)
	if err != nil {
		return "", nil, err
	}
	return r.Spec.renderSpiffeID(trustDomain, meta, environment)
}

func (i *MeshIdentity) GetSpiffeID(trustDomain string, meta model.ResourceMeta, environment config_core.EnvironmentType) (string, error) {
	spiffeID, _, err := i.renderSpiffeID(trustDomain, meta, environment)
	return spiffeID, err
}

func (i *MeshIdentity) renderSpiffeID(trustDomain string, meta model.ResourceMeta, environment config_core.EnvironmentType) (string, []string, error) {
	in := newSpiffeIDInputs(trustDomain, meta)
	path, err := renderTemplate(i.getSpiffeIDTemplate(environment), in.label, in)
	read := slices.Sorted(maps.Keys(in.read))
	if err != nil {
		return "", read, fmt.Errorf("failed to render SPIFFE ID template: %w", err)
	}
	spiffeID, err := spiffeid.FromString(path)
	if err != nil {
		return "", read, fmt.Errorf("failed to parse SPIFFE ID %q: %w", path, err)
	}
	return spiffeID.String(), read, nil
}

func lookupLabel(labels map[string]string, key string) (string, error) {
	val, ok := labels[key]
	if !ok {
		return "", errors.Errorf("label %s not found", key)
	}
	return val, nil
}

func renderTemplate(tmplStr string, label func(key string) (string, error), data any) (string, error) {
	tmpl, err := template.New("").Funcs(map[string]any{"label": label}).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed compiling Go template: %w", err)
	}

	var sb strings.Builder
	if err := tmpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("executing template failed: %w", err)
	}
	return sb.String(), nil
}

func (s *MeshIdentityStatus) IsInitialized() bool {
	for _, condition := range s.Conditions {
		if condition.Type == ReadyConditionType && condition.Status == kube_meta.ConditionTrue {
			return true
		}
	}
	return false
}

func (s *MeshIdentityStatus) IsPartiallyReady() bool {
	for _, condition := range s.Conditions {
		if condition.Type == ReadyConditionType && condition.Status == kube_meta.ConditionFalse && condition.Reason == "PartiallyReady" {
			return true
		}
	}
	return false
}
