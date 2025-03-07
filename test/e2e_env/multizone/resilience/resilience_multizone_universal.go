package resilience

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
)

func ResilienceMultizoneUniversal() {
	const clusterName1 = "kuma-res1"
	const clusterName2 = "kuma-res2"
	var global, zone1 Cluster

	BeforeEach(func() {
		// Global
		global = NewUniversalCluster(NewTestingT(), clusterName1, Silent)
		err := NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		zone1 = NewUniversalCluster(NewTestingT(), clusterName2, Silent)
		err = NewClusterSetup().
			Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEachFailure(func() {
		DebugUniversal(global, "default")
		DebugUniversal(zone1, "default")
	})

	E2EAfterEach(func() {
		err := zone1.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = zone1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma()
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should mark zone as offline after zone control-plane is killed forcefully", func() {
		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			if err != nil {
				return err
			}

			if !strings.Contains(output, "Online") {
				return errors.New("zone is not online")
			}

			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		Expect(zone1.(*UniversalCluster).Kill(AppModeCP, "kuma-cp run")).To(Succeed())

		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			if err != nil {
				return err
			}

			if !strings.Contains(output, "Offline") {
				return errors.New("zone is not offline")
			}

			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})
}
