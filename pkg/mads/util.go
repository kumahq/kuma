package mads

import (
	"fmt"
	"regexp"
	"strings"

	core_mesh "github.com/kumahq/kuma/v3/pkg/core/resources/apis/mesh"
)

func IndexMeshes(meshes []*core_mesh.MeshResource) map[string]*core_mesh.MeshResource {
	index := make(map[string]*core_mesh.MeshResource)
	for _, mesh := range meshes {
		index[mesh.Meta.GetName()] = mesh
	}
	return index
}

func MultiValue(values []string) string {
	// Although looks weird, it's actually a recommended way to represent multi-values in Prometheus.
	// It's meant to simplify greatly user-defined queries, e.g. just `,value,` instead of a full-featured regex.
	return "," + strings.Join(values, ",") + ","
}

func DataplaneLabels(dataplane *core_mesh.DataplaneResource) map[string]string {
	labels := map[string]string{}
	// first, we copy user-defined tags
	tags := dataplane.Spec.TagSet()
	for _, key := range tags.Keys() {
		values := tags.Values(key)
		value := ""
		if len(values) > 0 {
			value = values[0]
		}
		// while in general case a tag might have multiple values, we want to optimize for a single-value scenario
		labels[sanitizeLabelName(key)] = value
		// additionally, we also support a multi-value scenario by automatically pluralizing label name,
		// e.g. `service => services`, `version => versions`, etc.
		// if it happens that a user defined both `service` and `services` tags,
		// user-defined `services` tag will override auto-generated one (since keys are iterated in a sorted order)
		plural := fmt.Sprintf("%ss", key)
		labels[sanitizeLabelName(plural)] = MultiValue(values)
	}
	// then, we turn name extensions into labels
	for key, value := range dataplane.GetMeta().GetNameExtensions() {
		labels[sanitizeLabelName(key)] = value
	}

	return labels
}

func DataplaneAssignmentName(dataplane *core_mesh.DataplaneResource) string {
	// unique name, e.g. REST API uri
	return fmt.Sprintf("/meshes/%s/dataplanes/%s", dataplane.Meta.GetMesh(), dataplane.Meta.GetName())
}

var invalidLabelCharRE = regexp.MustCompile(`[^a-zA-Z0-9_]`)

// SanitizeLabelName replaces anything that doesn't match
// client_label.LabelNameRE with an underscore.
// taken from: https://github.com/prometheus/prometheus/blob/d437f0bb6b53ec8594a43b871f92252980b13ddd/util/strutil/strconv.go#L40-L47
func sanitizeLabelName(name string) string {
	return invalidLabelCharRE.ReplaceAllString(name, "_")
}
