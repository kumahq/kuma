package context_test

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cp "github.com/kumahq/kuma/v2/pkg/config/app/kuma-cp"
	config_core "github.com/kumahq/kuma/v2/pkg/config/core"
	config_store "github.com/kumahq/kuma/v2/pkg/config/core/resources/store"
	config_types "github.com/kumahq/kuma/v2/pkg/config/types"
	"github.com/kumahq/kuma/v2/pkg/core/resources/store"
	"github.com/kumahq/kuma/v2/pkg/kds/global"
	"github.com/kumahq/kuma/v2/pkg/kds/zone"
	"github.com/kumahq/kuma/v2/pkg/plugins/resources/memory"
	"github.com/kumahq/kuma/v2/pkg/test"
	"github.com/kumahq/kuma/v2/pkg/test/kds/setup"
	"github.com/kumahq/kuma/v2/pkg/test/matchers"
	test_store "github.com/kumahq/kuma/v2/pkg/test/store"
)

var _ = Describe("Full sync tests", func() {
	DescribeTable("Full sync tests", func(ctx SpecContext, folder string) {
		files, err := os.ReadDir(folder)
		Expect(err).ToNot(HaveOccurred())
		zones := make(map[string]store.ResourceStore)
		wg := sync.WaitGroup{}
		done := make(chan struct{})

		for _, file := range files {
			if before, ok := strings.CutSuffix(file.Name(), ".input.yaml"); ok {
				zoneName := before
				resourceStore := store.NewPaginationStore(memory.NewStore())
				fullPath := path.Join(folder, file.Name())
				Expect(test_store.LoadResourcesFromFile(ctx, resourceStore, fullPath)).To(Succeed())
				zones[zoneName] = resourceStore
			}
		}

		// Starts all the things

		globalStore := zones["global"]
		Expect(globalStore).ToNot(BeNil(), "global must be present")
		// start global
		cfg := kuma_cp.DefaultConfig()
		cfg.Store.Type = config_store.MemoryStore
		globalPort, err := test.GetFreePort()
		Expect(err).ToNot(HaveOccurred())
		cfg.Multizone.Global.KDS.GrpcPort = uint32(globalPort)
		cfg.Multizone.Global.KDS.TlsEnabled = false
		cfg.Multizone.Global.KDS.ZoneInsightFlushInterval = config_types.Duration{Duration: 100 * time.Millisecond}
		cfg.Mode = config_core.Global
		rt := setup.NewTestRuntime(ctx, cfg, globalStore)
		Expect(global.Setup(rt)).To(Succeed())
		wg.Go(func() {
			defer GinkgoRecover()
			Expect(rt.Start(done)).To(Succeed())
		})
		// Wait for the Global KDS listener to accept connections before
		// starting zones. Otherwise zones can race Global's bind and
		// hit "connection refused", which triggers the resilient
		// component's 5s backoff — roughly the same as the test's
		// 5s sync window, so the retry never lands in time.
		Eventually(func() error {
			c, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", globalPort), time.Second)
			if err != nil {
				return err
			}
			_ = c.Close()
			return nil
		}, "10s", "50ms").Should(Succeed())
		// start zones
		for zoneName, zoneStore := range zones {
			if zoneName == "global" {
				continue
			}
			cfg := kuma_cp.DefaultConfig()
			cfg.Store.Type = config_store.MemoryStore
			cfg.Mode = config_core.Zone
			cfg.Multizone.Zone.Name = zoneName
			cfg.Multizone.Zone.GlobalAddress = fmt.Sprintf("grpc://localhost:%d", globalPort)
			cfg.Multizone.Global.KDS.ZoneInsightFlushInterval = config_types.Duration{Duration: 100 * time.Millisecond}
			rt := setup.NewTestRuntime(ctx, cfg, zoneStore)
			Expect(zone.Setup(rt)).To(Succeed())
			wg.Go(func() {
				defer GinkgoRecover()
				Expect(rt.Start(done)).To(Succeed())
			})
		}

		// Wait until all stores converge to their expected golden state.
		// Using Eventually instead of a fixed sleep avoids flakes on
		// slow CI runners where 5s may not be enough for full sync.
		for zoneName, zoneStore := range zones {
			goldenFile := zoneName + ".golden.yaml"
			Eventually(func(g Gomega) {
				out, err := test_store.ExtractResources(ctx, zoneStore)
				g.Expect(err).To(Succeed())
				g.Expect(out).To(matchers.MatchGoldenEqual(folder, goldenFile), "zone %s", zoneName)
			}, "30s", "250ms").Should(Succeed())
		}

		close(done)
		wg.Wait()
	}, test.EntriesAsFolder("full_sync"))
})
