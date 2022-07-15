package main

import (
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("findCniConfFile", func() {
	It("should find conf file in a flat dir", func() {
		// given
		dir := path.Join("testdata", "find-conf-dir")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(Equal("10-flannel.conf"))
	})

	It("should find conflist file in a dir", func() {
		// given
		dir := path.Join("testdata", "find-conflist-dir")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(Not(HaveOccurred()))
		Expect(result).To(Equal("10-calico.conflist"))
	})

	It("should not find conf file in a nested dir", func() {
		// given
		dir := path.Join("testdata", "find-conf-dir-nested")

		// when
		result, err := findCniConfFile(dir)

		Expect(err).To(HaveOccurred())
		Expect(result).To(Equal(""))
	})
})
