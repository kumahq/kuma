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
	var global, zone1 Cluster

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)
		err = NewClusterSetup().
			Install(Kuma(core.Global)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		zone1 = clusters.GetCluster(Kuma2)
		err = NewClusterSetup().
			Install(Kuma(core.Zone, WithGlobalAddress(globalCP.GetKDSServerAddress()))).
			Setup(zone1)
		Expect(err).ToNot(HaveOccurred())
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

		_, _, err := zone1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

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
