package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

var _ = Describe("TrafficRouteHelper", func() {
	Context("GetSplitOrdered", func() {
		It("should return splits in consistent order", func() {
			tr1 := &TrafficRoute{Conf: &TrafficRoute_Conf{Split: []*TrafficRoute_Split{
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 10,
					},
					Destination: map[string]string{
						"key1": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 20,
					},
					Destination: map[string]string{
						"key2": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 1,
					},
					Destination: map[string]string{
						"key1": "value2",
						"key2": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 1,
					},
					Destination: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 10,
					},
					Destination: map[string]string{
						"key1": "value2",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 20,
					},
					Destination: map[string]string{
						"key2": "value1",
					},
				},
			}}}

			tr2 := &TrafficRoute{Conf: &TrafficRoute_Conf{Split: []*TrafficRoute_Split{
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 20,
					},
					Destination: map[string]string{
						"key2": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 20,
					},
					Destination: map[string]string{
						"key2": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 10,
					},
					Destination: map[string]string{
						"key1": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 1,
					},
					Destination: map[string]string{
						"key1": "value1",
						"key2": "value2",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 1,
					},
					Destination: map[string]string{
						"key1": "value2",
						"key2": "value1",
					},
				},
				{
					Weight: &wrapperspb.UInt32Value{
						Value: 10,
					},
					Destination: map[string]string{
						"key1": "value2",
					},
				},
			}}}

			sorted1 := tr1.GetConf().GetSplitOrdered()
			sorted2 := tr2.GetConf().GetSplitOrdered()

			for i := 0; i < len(tr1.GetConf().GetSplit()); i++ {
				Expect(sorted1[i].Weight).To(Equal(sorted2[i].Weight))
				Expect(sorted1[i].Destination).To(Equal(sorted2[i].Destination))
			}
		})
	})
})
