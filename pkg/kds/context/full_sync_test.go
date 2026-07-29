package context_test

import (
	"fmt"
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
		closeOnce := sync.Once{}
		DeferCleanup(func() {
			closeOnce.Do(func() { close(done) })
			wg.Wait()
		})

		for _, file := range files {
			if strings.HasSuffix(file.Name(), ".input.yaml") {
				zoneName := strings.TrimSuffix(file.Name(), ".input.yaml")
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer GinkgoRecover()
			Expect(rt.Start(done)).To(Succeed())
		}()
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
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer GinkgoRecover()
				Expect(rt.Start(done)).To(Succeed())
			}()
		}

		// A fixed sleep here flaked on slow CI runners before sync finished; poll instead.
		if os.Getenv("UPDATE_GOLDEN_FILES") != "" {
			time.Sleep(5 * time.Second)
		}
		for zoneName, zoneStore := range zones {
			goldenFile := zoneName + ".golden.yaml"
			Eventually(func(g Gomega) {
				out, err := test_store.ExtractResources(ctx, zoneStore)
				g.Expect(err).To(Succeed())
				g.Expect(out).To(matchers.MatchGoldenEqual(folder, goldenFile), "zone %s", zoneName)
			}, "30s", "250ms").Should(Succeed())
		}
	}, test.EntriesAsFolder("full_sync"))
})
