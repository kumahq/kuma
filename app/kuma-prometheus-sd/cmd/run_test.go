package cmd

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kuma_cmd "github.com/kumahq/kuma/pkg/cmd"
	"github.com/kumahq/kuma/pkg/test"
)

var _ = Describe("run", func() {

	var cancel func()
	var ctx context.Context
	opts := kuma_cmd.RunCmdOpts{
		SetupSignalHandler: func() context.Context {
			return ctx
		},
	}

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background())
	})

	XIt("should be possible to run `kuma-prometheus-sd run`", test.Within(15*time.Second, func() {
		// given
		cmd := NewRootCmd(opts)
		cmd.SetArgs([]string{"run"})

		// when
		By("starting the Kuma Prometheus SD")
		errCh := make(chan error)
		go func() {
			defer close(errCh)
			errCh <- cmd.Execute()
		}()

		// then
		By("waiting for Kuma Prometheus SD to become ready")

		// when
		By("signaling Kuma Prometheus SD to stop")
		cancel()

		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())
	}))
})
