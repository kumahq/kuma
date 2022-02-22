package v1alpha1_test

import (
	"time"

	envoy_resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/api/mesh/v1alpha1"
	system_proto "github.com/kumahq/kuma/api/system/v1alpha1"
	util_proto "github.com/kumahq/kuma/pkg/util/proto"
)

var _ = Describe("DataplaneHelpers", func() {

	Describe("DataplaneInsight", func() {

		var status *DataplaneInsight
		var t1, t2, t3 time.Time

		BeforeEach(func() {
			status = &DataplaneInsight{}
			t1, _ = time.Parse(time.RFC3339, "2017-07-17T17:07:47+00:00")
			t2, _ = time.Parse(time.RFC3339, "2018-08-18T18:08:48+00:00")
			t3, _ = time.Parse(time.RFC3339, "2019-09-19T19:09:49+00:00")
		})

		Describe("UpdateSubscription()", func() {

			It("should add new subscriptions", func() {
				// given
				subscription := &DiscoverySubscription{
					Id:                     "1",
					ControlPlaneInstanceId: "node-001",
					Status:                 NewSubscriptionStatus(),
				}

				// when
				Expect(status.UpdateSubscription(subscription)).To(Succeed())

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                subscriptions:
                - controlPlaneInstanceId: node-001
                  id: "1"
                  status:
                    cds: {}
                    eds: {}
                    lds: {}
                    rds: {}
                    total: {}
`))
			})

			It("should replace existing subscriptions", func() {
				// setup
				status.Subscriptions = []*DiscoverySubscription{
					{
						Id:                     "1",
						ControlPlaneInstanceId: "node-001",
						Status:                 NewSubscriptionStatus(),
					},
					{
						Id:                     "2",
						ControlPlaneInstanceId: "node-002",
						Status:                 NewSubscriptionStatus(),
					},
				}

				// given
				subscription := &DiscoverySubscription{
					Id:                     "1",
					ControlPlaneInstanceId: "node-003",
					Status:                 NewSubscriptionStatus(),
				}

				// when
				Expect(status.UpdateSubscription(subscription)).To(Succeed())

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                subscriptions:
                - controlPlaneInstanceId: node-003
                  id: "1"
                  status:
                    cds: {}
                    eds: {}
                    lds: {}
                    rds: {}
                    total: {}
                - controlPlaneInstanceId: node-002
                  id: "2"
                  status:
                    cds: {}
                    eds: {}
                    lds: {}
                    rds: {}
                    total: {}
`))
			})

			It("should leave subscriptions in a valid state", func() {
				// given
				dataplaneInsight := &DataplaneInsight{
					Subscriptions: []*DiscoverySubscription{
						{
							Id:             "1",
							ConnectTime:    util_proto.MustTimestampProto(t1),
							DisconnectTime: util_proto.MustTimestampProto(t1.Add(1 * time.Hour)),
						},
						{
							Id:          "2",
							ConnectTime: util_proto.MustTimestampProto(t1.Add(2 * time.Hour)),
						},
					},
				}

				// when
				Expect(dataplaneInsight.UpdateSubscription(&DiscoverySubscription{
					Id:          "3",
					ConnectTime: util_proto.MustTimestampProto(t1.Add(3 * time.Hour)),
				})).To(Succeed())

				// then
				_, subscription := dataplaneInsight.GetSubscription("2")
				Expect(subscription.DisconnectTime).ToNot(BeNil())
			})

			It("should return error for wrong subscription type", func() {
				// given
				dataplaneInsight := &DataplaneInsight{
					Subscriptions: []*DiscoverySubscription{
						{
							Id:             "1",
							ConnectTime:    util_proto.MustTimestampProto(t1),
							DisconnectTime: util_proto.MustTimestampProto(t1.Add(1 * time.Hour)),
						},
						{
							Id:          "2",
							ConnectTime: util_proto.MustTimestampProto(t1.Add(2 * time.Hour)),
						},
					},
				}

				// when
				err := dataplaneInsight.UpdateSubscription(&system_proto.KDSSubscription{})

				// then
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("invalid type *v1alpha1.KDSSubscription for DataplaneInsight"))
			})
		})

		Describe("GetLastSubscription()", func() {

			It("should return `nil` when there are no subscriptions", func() {
				// given
				status.Subscriptions = nil

				// when
				subscription := status.GetLastSubscription()

				// then
				Expect(subscription).To(BeNil())
			})

			It("should return last subscription", func() {
				// given
				status.Subscriptions = []*DiscoverySubscription{
					{
						Id:          "1",
						ConnectTime: util_proto.MustTimestampProto(t1),
					},
					{
						Id:          "3",
						ConnectTime: util_proto.MustTimestampProto(t3),
					},
					{
						Id:          "2",
						ConnectTime: util_proto.MustTimestampProto(t2),
					},
				}

				// when
				subscription := status.GetLastSubscription()

				// then
				Expect(subscription).To(BeIdenticalTo(status.Subscriptions[2]))
				Expect(subscription.GetId()).To(Equal("2"))
			})
		})

		Describe("Sum()", func() {

			It("should return `0` when there are no subscriptions", func() {
				// given
				status.Subscriptions = nil

				// when
				sum := status.Sum(func(s *DiscoverySubscription) uint64 {
					return s.Status.Total.ResponsesSent
				})

				// then
				Expect(sum).To(Equal(uint64(0)))
			})

			It("should return sum across all subscriptions", func() {
				// given
				status.Subscriptions = []*DiscoverySubscription{
					{
						Id: "1",
						Status: &DiscoverySubscriptionStatus{
							Total: &DiscoveryServiceStats{
								ResponsesSent: 1,
							},
						},
					},
					{
						Id: "2",
						Status: &DiscoverySubscriptionStatus{
							Total: &DiscoveryServiceStats{
								ResponsesSent: 2,
							},
						},
					},
				}

				// when
				sum := status.Sum(func(s *DiscoverySubscription) uint64 {
					return s.Status.Total.ResponsesSent
				})

				// then
				Expect(sum).To(Equal(uint64(3)))
			})
		})
	})

	Describe("DiscoverySubscriptionStatus", func() {

		var status *DiscoverySubscriptionStatus

		BeforeEach(func() {
			status = NewSubscriptionStatus()
		})

		Describe("StatsOf()", func() {

			It("should support CDS", func() {

				// when
				status.StatsOf(envoy_resource.ClusterType).ResponsesSent = 1

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                cds:
                  responsesSent: "1"
                eds: {}
                lds: {}
                rds: {}
                total: {}
`))
			})

			It("should support EDS", func() {

				// when
				status.StatsOf(envoy_resource.EndpointType).ResponsesSent = 1

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                cds: {}
                eds:
                  responsesSent: "1"
                lds: {}
                rds: {}
                total: {}
`))
			})

			It("should support LDS", func() {

				// when
				status.StatsOf(envoy_resource.ListenerType).ResponsesSent = 1

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                cds: {}
                eds: {}
                lds:
                  responsesSent: "1"
                rds: {}
                total: {}
`))
			})

			It("should support RDS", func() {

				// when
				status.StatsOf(envoy_resource.RouteType).ResponsesSent = 1

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                cds: {}
                eds: {}
                lds: {}
                rds:
                  responsesSent: "1"
                total: {}
`))
			})

			It("should not fail on unknown xDS", func() {
				// when
				status.StatsOf(envoy_resource.SecretType).ResponsesSent = 1

				// then
				Expect(util_proto.ToYAML(status)).To(MatchYAML(`
                cds: {}
                eds: {}
                lds: {}
                rds: {}
                total: {}
`))
			})
		})
	})

	type testCase struct {
		inputVersion    string
		expectedVersion string
		expectedLabel   string
	}
	DescribeTable("Envoy.ParseVersion",
		func(given testCase) {
			actualVersion, actualLabel := (&EnvoyVersion{
				Version: given.inputVersion,
			}).ParseVersion()
			Expect(actualVersion).To(Equal(given.expectedVersion))
			Expect(actualLabel).To(Equal(given.expectedLabel))
		},
		Entry("empty", testCase{
			inputVersion:    "",
			expectedVersion: "",
			expectedLabel:   "",
		}),
		Entry("no label", testCase{
			inputVersion:    "1.20.0",
			expectedVersion: "1.20.0",
			expectedLabel:   "",
		}),
		Entry("simple label", testCase{
			inputVersion:    "1.20.0-dev",
			expectedVersion: "1.20.0",
			expectedLabel:   "dev",
		}),
		Entry("label with dashes", testCase{
			inputVersion:    "1.20.0-super-dev",
			expectedVersion: "1.20.0",
			expectedLabel:   "super-dev",
		}),
	)
})
