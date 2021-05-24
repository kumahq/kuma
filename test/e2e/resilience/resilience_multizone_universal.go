package resilience

import (
	"errors"
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/kumahq/kuma/test/framework"
)

func ResilienceMultizoneUniversal() {
	var global, remote_1 Cluster
	var optsGlobal, optsRemote1 []DeployOptionsFunc

	BeforeEach(func() {
		clusters, err := NewUniversalClusters(
			[]string{Kuma1, Kuma2},
			Silent)
		Expect(err).ToNot(HaveOccurred())

		// Global
		global = clusters.GetCluster(Kuma1)
		optsGlobal = []DeployOptionsFunc{}
		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())
		err = global.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())

		globalCP := global.GetKuma()

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)
		optsRemote1 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.VerifyKuma()
		Expect(err).ToNot(HaveOccurred())
	})

	E2EAfterEach(func() {
		err := remote_1.DeleteKuma(optsRemote1...)
		Expect(err).ToNot(HaveOccurred())
		err = remote_1.DismissCluster()
		Expect(err).ToNot(HaveOccurred())

		err = global.DeleteKuma(optsGlobal...)
		Expect(err).ToNot(HaveOccurred())
		err = global.DismissCluster()
		Expect(err).ToNot(HaveOccurred())
	})

	It("should mark zone as offline after remote control-plane is killed forcefully", func() {
		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			if err != nil {
				return err
			}

			if !strings.Contains(output, "Online") {
				return errors.New("remote zone is not online")
			}

			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

		_, _, err := remote_1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() error {
			output, err := global.GetKumactlOptions().RunKumactlAndGetOutput("inspect", "zones")
			if err != nil {
				return err
			}

			if !strings.Contains(output, "Offline") {
				return errors.New("remote zone is not offline")
			}

			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())
	})
}
