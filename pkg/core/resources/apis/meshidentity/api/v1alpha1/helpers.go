package v1alpha1

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	mesh_proto "github.com/kumahq/kuma/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/pkg/core/resources/model"
	"github.com/kumahq/kuma/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/pkg/util/pointer"
	"github.com/pkg/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultTrustDomainTemplate = "{{ .Mesh }}.{{ .Zone }}.mesh.local"
	defaultPathTemplate        = "/ns/{{ .Namespace }}/sa/{{ .ServiceAccount }}"
)

func Matched(
	labels map[string]string,
	meshIdentities []*MeshIdentityResource,
) (*MeshIdentityResource, bool) {
	type scoredMatch struct {
		matchCount int
		name       string
		policy     *MeshIdentityResource
	}

	var matches []scoredMatch

	for _, mi := range meshIdentities {
		if mi.Spec.Selector == nil || mi.Spec.Selector.Dataplane == nil || !mi.Spec.Selector.Dataplane.Matches(labels) {
			continue
		}

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

func (i *MeshIdentity) getSpiffeIDTemplate() string {
	builder := strings.Builder{}
	builder.WriteString("spiffe://")
	if i.SpiffeID != nil {
		builder.WriteString(pointer.DerefOr(i.SpiffeID.Path, defaultPathTemplate))
	} else {
		builder.WriteString("{{ .TrustDomain }}")
		builder.WriteString(defaultPathTemplate)
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

func (i *MeshIdentity) GetSpiffeID(trustDomain string, meta model.ResourceMeta) (string, error) {
	spiffeIDTemplate := i.getSpiffeIDTemplate()

	data := struct {
		TrustDomain    string
		Namespace      string
		ServiceAccount string
	}{
		TrustDomain:    trustDomain,
		Namespace:      meta.GetLabels()[mesh_proto.KubeNamespaceTag],
		ServiceAccount: meta.GetLabels()[metadata.KumaServiceAccount],
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
