package v1alpha1_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	zoneinsight_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zoneinsight/api/v1alpha1"
)

// These fixtures are the bytes a 2.14 control plane wrote, captured from its store rather
// than written by hand, so the test fails if the Go structs drift from what protobuf
// produced. A 3.0 control plane has to read them, and has to write bytes 2.14 still reads,
// because a virtual control plane can be rolled back onto the same database.
var _ = Describe("ZoneInsight storage compatibility", func() {
	entries, err := os.ReadDir("testdata/compat")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		name := entry.Name()

		It("reads and rewrites "+name+" without changing a byte", func() {
			stored, err := os.ReadFile(filepath.Join("testdata/compat", name))
			Expect(err).ToNot(HaveOccurred())

			insight := &zoneinsight_api.ZoneInsight{}
			Expect(json.Unmarshal(stored, insight)).To(Succeed())

			rewritten, err := json.Marshal(insight)
			Expect(err).ToNot(HaveOccurred())

			// Compared as parsed JSON: key order is not part of the contract, values are.
			var was, is any
			Expect(json.Unmarshal(stored, &was)).To(Succeed())
			Expect(json.Unmarshal(rewritten, &is)).To(Succeed())
			Expect(is).To(Equal(was))
		})
	}

	It("writes 64 bit counters as strings, the way protobuf did", func() {
		out, err := json.Marshal(&zoneinsight_api.ZoneInsight{
			Subscriptions: []*zoneinsight_api.KDSSubscription{{
				ID: "1",
				Status: &zoneinsight_api.KDSSubscriptionStatus{
					Total: &zoneinsight_api.KDSServiceStats{ResponsesSent: 42},
				},
			}},
		})
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).To(ContainSubstring(`"responsesSent":"42"`))
	})

	It("writes timestamps in UTC whatever zone the control plane runs in", func() {
		kolkata, err := time.LoadLocation("Asia/Kolkata")
		Expect(err).ToNot(HaveOccurred())
		instant := time.Date(2026, 9, 8, 10, 30, 0, 0, time.UTC)

		utc, err := json.Marshal(zoneinsight_api.NewTime(instant))
		Expect(err).ToNot(HaveOccurred())
		shifted, err := json.Marshal(zoneinsight_api.NewTime(instant.In(kolkata)))
		Expect(err).ToNot(HaveOccurred())

		Expect(string(shifted)).To(Equal(string(utc)))
		Expect(string(utc)).To(Equal(`"2026-09-08T10:30:00Z"`))
	})

	It("keeps an unset enabled flag distinct from an explicit false", func() {
		var unset zoneinsight_api.KDSSubscription
		Expect(json.Unmarshal([]byte(`{"id":"1"}`), &unset)).To(Succeed())
		Expect(unset.AuthTokenProvided).To(BeFalse())

		out, err := json.Marshal(&unset)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(out)).ToNot(ContainSubstring("authTokenProvided"))
	})
})
