package v1alpha1

import (
	"fmt"
	"sort"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	kube_meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	mesh_proto "github.com/kumahq/kuma/v2/api/mesh/v1alpha1"
	"github.com/kumahq/kuma/v2/pkg/core/resources/model"
	"github.com/kumahq/kuma/v2/pkg/plugins/runtime/k8s/metadata"
	"github.com/kumahq/kuma/v2/pkg/util/pointer"
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

<<<<<<< HEAD
func (i *MeshIdentity) getSpiffeIDTemplate() string {
	builder := strings.Builder{}
	builder.WriteString("spiffe://")
	builder.WriteString("{{ .TrustDomain }}")
	if i.SpiffeID != nil {
		builder.WriteString(pointer.DerefOr(i.SpiffeID.Path, defaultPathTemplate))
	} else {
		builder.WriteString(defaultPathTemplate)
	}
=======
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
>>>>>>> 149a59a47d (fix(meshidentity): env-aware UsesWorkloadLabel (#16356))
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
<<<<<<< HEAD
=======

func (s *MeshIdentityStatus) IsPartiallyReady() bool {
	for _, condition := range s.Conditions {
		if condition.Type == ReadyConditionType && condition.Status == kube_meta.ConditionFalse && condition.Reason == "PartiallyReady" {
			return true
		}
	}
	return false
}

var (
	workloadLabelRegex       = regexp.MustCompile(`\{\{\s*label\s+"kuma\.io/workload"\s*\}\}`)
	workloadPlaceholderRegex = regexp.MustCompile(`\{\{\s*\.Workload\s*\}\}`)
)

// UsesWorkloadLabel checks if this MeshIdentity's SPIFFE ID path template contains
// the workload reference in the form of {{ label "kuma.io/workload" }} or {{ .Workload }}.
// The environment is taken into account because the default Universal path template
// references .Workload even when no custom path is configured.
func (i *MeshIdentity) UsesWorkloadLabel(env config_core.EnvironmentType) bool {
	path := i.SpiffeIDPathTemplate(env)
	return workloadLabelRegex.MatchString(path) || workloadPlaceholderRegex.MatchString(path)
}
>>>>>>> 149a59a47d (fix(meshidentity): env-aware UsesWorkloadLabel (#16356))
