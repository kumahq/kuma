package xds

import (
	"fmt"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/discovery/targetgroup"
	"github.com/prometheus/prometheus/util/strutil"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
)

type Converter struct{}

func (c Converter) ConvertAll(assignments []*observability_proto.MonitoringAssignment) []*targetgroup.Group {
	var groups []*targetgroup.Group
	for _, assignment := range assignments {
		groups = append(groups, c.Convert(assignment)...)
	}
	return groups
}

func (c Converter) Convert(assignment *observability_proto.MonitoringAssignment) []*targetgroup.Group {
	var groups []*targetgroup.Group
	commonLabels := c.convertLabels(assignment.Labels)
	for i, target := range assignment.Targets {
		targetLabels := c.convertLabels(target.Labels)
		allLabels := commonLabels.Clone().Merge(targetLabels)

		address := allLabels[model.AddressLabel]
		delete(allLabels, model.AddressLabel)

		group := &targetgroup.Group{
			Source: c.subSourceName(assignment.Name, i),
			Targets: []model.LabelSet{{
				model.AddressLabel: address,
			}},
			Labels: allLabels,
		}
		groups = append(groups, group)
	}
	return groups
}

func (c Converter) convertLabels(labels map[string]string) model.LabelSet {
	labelSet := model.LabelSet{}
	for key, value := range labels {
		name := strutil.SanitizeLabelName(key)
		labelSet[model.LabelName(name)] = model.LabelValue(value)
	}
	return labelSet
}

func (c Converter) subSourceName(source string, i int) string {
	return fmt.Sprintf("%s/%d", source, i)
}
