package v1alpha1

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
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

func (i *MeshIdentity) getSpiffeIDTemplate(env config_core.EnvironmentType) string {
	var defaultSpiffeIDPathTemplate string
	switch env {
	case config_core.KubernetesEnvironment:
		defaultSpiffeIDPathTemplate = defaultK8sSpiffeIDPathTemplate
	case config_core.UniversalEnvironment:
		defaultSpiffeIDPathTemplate = defaultUniversalSpiffeIDPathTemplate
	}
	builder := strings.Builder{}
	builder.WriteString("spiffe://")
	builder.WriteString("{{ .TrustDomain }}")
	if i.SpiffeID != nil {
		builder.WriteString(pointer.DerefOr(i.SpiffeID.Path, defaultSpiffeIDPathTemplate))
	} else {
		builder.WriteString(defaultSpiffeIDPathTemplate)
	}
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

	return renderTemplate(trustDomainTmpl, meta, data)
}

func (i *MeshIdentity) GetSpiffeID(trustDomain string, meta model.ResourceMeta, environment config_core.EnvironmentType) (string, error) {
	spiffeIDTemplate := i.getSpiffeIDTemplate(environment)

	data := struct {
		TrustDomain    string
		Namespace      string
		ServiceAccount string
		Workload       string
	}{
		TrustDomain:    trustDomain,
		Namespace:      meta.GetLabels()[mesh_proto.KubeNamespaceTag],
		ServiceAccount: meta.GetLabels()[metadata.KumaServiceAccount],
		Workload:       meta.GetLabels()[metadata.KumaWorkload],
	}

	return renderTemplate(spiffeIDTemplate, meta, data)
}

func renderTemplate(tmplStr string, meta model.ResourceMeta, data any) (string, error) {
	tmpl, err := template.New("").Funcs(map[string]any{
		"label": func(key string) (string, error) {
			val, ok := meta.GetLabels()[key]
			if !ok {
				return "", errors.Errorf("label %s not found", key)
			}
			return val, nil
		},
	}).Parse(tmplStr)
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
