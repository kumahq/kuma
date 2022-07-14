package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("validation", func() {
	Context("isValidConflistFile", func() {
		It("should return no error when a file is a valid conflist file", func() {
			err := isValidConflistFile("testdata/10-calico.conflist")

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should return error when a file is an invalid conflist file", func() {
			err := isValidConflistFile("testdata/10-flannel.conf")

			Expect(err).To(HaveOccurred())
		})
	})

	Context("isValidConfFile", func() {
		It("should return no error when a file is a valid conf file", func() {
			err := isValidConfFile("testdata/10-flannel.conf")

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should return false when a file is an invalid conf file", func() {
			err := isValidConfFile("testdata/10-calico.conflist")

			Expect(err).To(HaveOccurred())
		})
	})

	Context("checkInstall", func() {
		It("should not return error when a file is a conflist file with kuma-cni installed", func() {
			err := checkInstall("testdata/10-flannel.conf.golden", true)

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should return true when a file is a conf file with kuma-cni", func() {
			err := checkInstall("testdata/10-kuma-cni.conf", false)

			Expect(err).To(Not(HaveOccurred()))
		})

		It("should return false when a file does not have kuma cni installed", func() {
			err := checkInstall("testdata/10-flannel.conf", false)

			Expect(err).To(HaveOccurred())
		})
	})
})
