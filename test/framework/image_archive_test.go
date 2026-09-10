package framework

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("dockerSaveArgs", func() {
	It("pins the archive when a platform is required", func() {
		// Given a daemon whose store needs the archive pinned
		// When the save arguments are built
		args := dockerSaveArgs("/tmp/images.tar", "linux/amd64", []string{"a:1", "b:2"})

		// Then the archive is restricted to that platform
		Expect(args).To(Equal([]string{
			"image", "save",
			"--platform", "linux/amd64",
			"-o", "/tmp/images.tar",
			"a:1", "b:2",
		}))
	})

	It("passes no flag when no platform is required", func() {
		// Given a daemon that holds one variant per image already
		// When the save arguments are built
		args := dockerSaveArgs("/tmp/images.tar", "", []string{"a:1"})

		// Then nothing an older CLI would reject is passed
		Expect(args).To(Equal([]string{"image", "save", "-o", "/tmp/images.tar", "a:1"}))
	})
})
