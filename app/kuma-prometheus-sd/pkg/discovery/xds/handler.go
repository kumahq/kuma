package xds

import (
	"github.com/prometheus/prometheus/discovery/targetgroup"

	observability_proto "github.com/Kong/kuma/api/observability/v1alpha1"
)

type sourceList = map[string]bool

type Handler struct {
	converter     Converter
	oldSourceList sourceList
}

func (h *Handler) Handle(assignments []*observability_proto.MonitoringAssignment, ch chan<- []*targetgroup.Group) {
	newGroups := h.converter.ConvertAll(assignments)
	newSourceList := h.buildSourceList(newGroups)
	removedGroups := h.buildRemovedGroups(newSourceList)
	allGroups := append(newGroups, removedGroups...)
	ch <- allGroups
	h.oldSourceList = newSourceList
}

func (h *Handler) buildSourceList(newGroups []*targetgroup.Group) sourceList {
	newSourceList := sourceList{}
	for _, group := range newGroups {
		newSourceList[group.Source] = true
	}
	return newSourceList
}

func (h *Handler) buildRemovedGroups(newSourceList sourceList) []*targetgroup.Group {
	// when targetGroup disappears, we should send an update with an empty targetList
	var deletedGroups []*targetgroup.Group
	for name := range h.oldSourceList {
		if !newSourceList[name] {
			deletedGroups = append(deletedGroups, &targetgroup.Group{
				Source: name,
			})
		}
	}
	return deletedGroups
}
