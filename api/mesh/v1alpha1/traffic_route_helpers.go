package v1alpha1

import (
	"fmt"
	"sort"
	"strings"

	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

// GetSplitWithDestination returns unified list of split regardless if split or destination is used
// Destination is a syntax sugar over single split with weight of 1.
func (x *TrafficRoute_Conf) GetSplitWithDestination() []*TrafficRoute_Split {
	if len(x.GetDestination()) > 0 {
		return []*TrafficRoute_Split{
			{
				Weight:      util_proto.UInt32(1),
				Destination: x.GetDestination(),
			},
		}
	}
	return x.GetSplit()
}

func (x *TrafficRoute_Http) GetSplitWithDestination() []*TrafficRoute_Split {
	if len(x.GetDestination()) > 0 {
		return []*TrafficRoute_Split{
			{
				Weight:      util_proto.UInt32(1),
				Destination: x.GetDestination(),
			},
		}
	}
	return x.GetSplit()
}

func (x *TrafficRoute_Conf) GetSplitOrdered() []*TrafficRoute_Split {
	c := make([]*TrafficRoute_Split, len(x.GetSplitWithDestination()))
	copy(c, x.GetSplitWithDestination())
	sort.Stable(SortedSplit(c))
	return c
}

type SortedSplit []*TrafficRoute_Split

func (s SortedSplit) Len() int { return len(s) }
func (s SortedSplit) Less(i, j int) bool {
	if s[i].GetWeight().GetValue() != s[j].GetWeight().GetValue() {
		return s[i].GetWeight().GetValue() < s[j].GetWeight().GetValue()
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
