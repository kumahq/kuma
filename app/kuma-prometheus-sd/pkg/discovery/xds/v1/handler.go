package v1

import (
	"github.com/prometheus/prometheus/discovery/targetgroup"

	observability_v1 "github.com/kumahq/kuma/api/observability/v1"
)

type sourceList = map[string]bool

type Handler struct {
	converter     Converter
	oldSourceList sourceList
}

func (h *Handler) Handle(assignments []*observability_v1.MonitoringAssignment, ch chan<- []*targetgroup.Group) {
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
