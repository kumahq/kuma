package resilience

import (
	"strings"

	"github.com/kumahq/kuma/pkg/config/core"
	. "github.com/kumahq/kuma/test/framework"
	"github.com/kumahq/kuma/test/framework/deployments/postgres"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
)

func ResilienceMultizoneUniversalPostgres() {
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
			Install(postgres.Install(Kuma1)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		optsGlobal = []DeployOptionsFunc{
			WithPostgres(postgres.From(global, Kuma1).GetEnvVars()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Global, optsGlobal...)).
			Setup(global)
		Expect(err).ToNot(HaveOccurred())

		Expect(global.VerifyKuma()).Should(Succeed())

		globalCP := global.GetKuma()

		// Cluster 1
		remote_1 = clusters.GetCluster(Kuma2)

		err = NewClusterSetup().
			Install(postgres.Install(Kuma2)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())

		optsRemote1 = []DeployOptionsFunc{
			WithGlobalAddress(globalCP.GetKDSServerAddress()),
			WithPostgres(postgres.From(remote_1, Kuma2).GetEnvVars()),
		}

		err = NewClusterSetup().
			Install(Kuma(core.Remote, optsRemote1...)).
			Setup(remote_1)
		Expect(err).ToNot(HaveOccurred())

		Expect(remote_1.VerifyKuma()).Should(Succeed())
	})

	E2EAfterEach(func() {
		// remote
		Expect(remote_1.DeleteKuma(optsRemote1...)).Should(Succeed())
		Expect(remote_1.DismissCluster()).Should(Succeed())

		// global
		Expect(global.DeleteKuma(optsGlobal...)).Should(Succeed())
		Expect(global.DismissCluster()).Should(Succeed())
	})

	It("should mark zone as offline after remote control-plane is killed forcefully when global control-plane is down", func() {
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

		g, ok := global.(*UniversalCluster)
		Expect(ok).To(BeTrue())

		kumaCP := g.GetApp(AppKumaCP)
		Expect(kumaCP).ToNot(BeNil())

		_, _, err := global.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		_, _, err = remote_1.Exec("", "", AppModeCP, "pkill", "-9", "kuma-cp")
		Expect(err).ToNot(HaveOccurred())

		err = kumaCP.ReStart()
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() error {
			if err := global.VerifyKuma(); err != nil {
				return err
			}

			return nil
		}, "30s", "1s").ShouldNot(HaveOccurred())

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
