package v1alpha1

import (
	"fmt"
	"sort"
	"strings"
)

func (x *TrafficRoute_Conf) GetSplitOrdered() []*TrafficRoute_Split {
	c := make([]*TrafficRoute_Split, len(x.GetSplit()))
	copy(c, x.GetSplit())
	sort.Stable(SortedSplit(c))
	return c
}

type SortedSplit []*TrafficRoute_Split

func (s SortedSplit) Len() int { return len(s) }
func (s SortedSplit) Less(i, j int) bool {
	if s[i].Weight != s[j].Weight {
		return s[i].Weight < s[j].Weight
	}
	return less(s[i].Destination, s[j].Destination)
}
func (s SortedSplit) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func less(m1, m2 map[string]string) bool {
	return destination(m1).String() < destination(m2).String()
}

type destination map[string]string

func (d destination) String() string {
	pairs := make([]string, 0, len(d))
	for k, v := range d {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "&")
}
