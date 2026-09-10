package v1alpha1_test

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	zone_api "github.com/kumahq/kuma/v3/pkg/core/resources/apis/zone/api/v1alpha1"
	"github.com/kumahq/kuma/v3/pkg/util/pointer"
)

// The bytes below are what a 2.14 control plane wrote for each state of the flag. A 3.0
// control plane has to read them and write them back unchanged: the Kubernetes admission
// path re-serializes a resource on every write, so any drift here would rewrite every
// stored Zone and would reach an older control plane on rollback.
var _ = Describe("Zone storage compatibility", func() {
	DescribeTable("round trips the bytes 2.14 wrote",
		func(stored string, expected *bool) {
			zone := &zone_api.Zone{}
			Expect(json.Unmarshal([]byte(stored), zone)).To(Succeed())
			Expect(zone.Enabled).To(Equal(expected))

			rewritten, err := json.Marshal(zone)
			Expect(err).ToNot(HaveOccurred())
			Expect(string(rewritten)).To(Equal(stored))
		},
		Entry("enabled", `{"enabled":true}`, pointer.To(true)),
		Entry("explicitly disabled", `{"enabled":false}`, pointer.To(false)),
		Entry("never set", `{}`, nil),
	)

	It("treats a flag that was never set as enabled, the way the wrapper did", func() {
		Expect((&zone_api.Zone{}).IsEnabled()).To(BeTrue())
		Expect((&zone_api.Zone{Enabled: pointer.To(false)}).IsEnabled()).To(BeFalse())
		Expect((&zone_api.Zone{Enabled: pointer.To(true)}).IsEnabled()).To(BeTrue())
	})
})
