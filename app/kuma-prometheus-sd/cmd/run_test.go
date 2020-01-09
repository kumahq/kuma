package cmd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("run", func() {

	var backupSetupSignalHandler func() <-chan struct{}

	BeforeEach(func() {
		backupSetupSignalHandler = setupSignalHandler
	})

	AfterEach(func() {
		setupSignalHandler = backupSetupSignalHandler
	})

	var stopCh chan struct{}

	BeforeEach(func() {
		stopCh = make(chan struct{})

		setupSignalHandler = func() <-chan struct{} {
			return stopCh
		}
	})

	XIt("should be possible to run `kuma-prometheus-sd run`", func(done Done) {
		// given
		cmd := NewRootCmd()
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
		// TODO(yskopets)

		// when
		By("signalling Kuma Prometheus SD to stop")
		close(stopCh)

		// then
		err := <-errCh
		Expect(err).ToNot(HaveOccurred())

		// complete
		close(done)
	}, 15)
})
