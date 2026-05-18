package framework

import (
	"regexp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UniversalApp", func() {
	Describe("mainAppProcessRegex", func() {
		It("matches path-prefixed command", func() {
			app := &UniversalApp{}
			regex := regexp.MustCompile(app.mainAppProcessRegex("kuma-cp run --config-file /kuma/kuma-cp.conf"))

			Expect(regex.MatchString("/usr/bin/kuma-cp run --config-file /kuma/kuma-cp.conf")).To(BeTrue())
			Expect(regex.MatchString("kuma-cp run --config-file /kuma/kuma-cp.conf")).To(BeTrue())
			Expect(regex.MatchString("/usr/bin/not-kuma-cp run --config-file /kuma/kuma-cp.conf")).To(BeFalse())
		})
	})
})
