package cmd

import (
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("drain", func() {
	var procRoot string
	var signaled map[int]syscall.Signal

	writeProc := func(pid int, comm string) {
		dir := filepath.Join(procRoot, strconv.Itoa(pid))
		Expect(os.MkdirAll(dir, 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(dir, "comm"), []byte(comm+"\n"), 0o600)).To(Succeed())
	}

	run := func() error {
		cmd := newDrainCmd(procRoot, func(pid int, sig syscall.Signal) error {
			signaled[pid] = sig
			return nil
		})
		cmd.SetArgs([]string{})
		cmd.SetOut(GinkgoWriter)
		return cmd.Execute()
	}

	BeforeEach(func() {
		procRoot = GinkgoT().TempDir()
		signaled = map[int]syscall.Signal{}
	})

	It("should send SIGUSR2 to other kuma-dp processes only", func() {
		// given
		writeProc(1, "kuma-dp")
		writeProc(7, "envoy")
		writeProc(os.Getpid(), "kuma-dp")
		Expect(os.MkdirAll(filepath.Join(procRoot, "self"), 0o755)).To(Succeed())

		// when
		err := run()

		// then
		Expect(err).ToNot(HaveOccurred())
		Expect(signaled).To(Equal(map[int]syscall.Signal{1: syscall.SIGUSR2}))
	})

	It("should fail when no kuma-dp process is running", func() {
		// given
		writeProc(7, "envoy")

		// when
		err := run()

		// then
		Expect(err).To(MatchError(ContainSubstring("no kuma-dp process found")))
		Expect(signaled).To(BeEmpty())
	})
})
