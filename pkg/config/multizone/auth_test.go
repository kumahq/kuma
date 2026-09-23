package multizone_test

import (
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kumahq/kuma/v3/pkg/config/multizone"
)

var _ = Describe("KDSServerAuthConfig", func() {
	DescribeTable("should validate the auth type",
		func(authType multizone.KDSAuthType, errSubstring string) {
			err := multizone.KDSServerAuthConfig{Type: authType}.Validate()

			if errSubstring == "" {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err).To(MatchError(ContainSubstring(errSubstring)))
			}
		},
		Entry("none", multizone.KDSAuthNone, ""),
		Entry("zoneToken", multizone.KDSAuthZoneToken, ""),
		Entry("a type a distribution registers", multizone.KDSAuthType("custom"), ""),
		Entry("empty", multizone.KDSAuthType(""), ".Type cannot be empty"),
	)
})

var _ = Describe("KDSClientAuthConfig", func() {
	type tokenFiles struct {
		token      string
		empty      string
		dotted     string
		traversing string
	}
	var files tokenFiles

	writeToken := func(path string, token string) string {
		Expect(os.WriteFile(path, []byte(token), 0o600)).To(Succeed())
		return path
	}

	BeforeEach(func() {
		tokenDir := GinkgoT().TempDir()
		files.token = writeToken(filepath.Join(tokenDir, "token"), "file-token\n")
		// built by hand, filepath.Join would clean the ".." away before LoadToken sees it
		files.traversing = strings.Join([]string{tokenDir, "..", filepath.Base(tokenDir), "token"}, string(filepath.Separator))

		files.empty = writeToken(filepath.Join(GinkgoT().TempDir(), "empty"), "\n")

		dottedDir := filepath.Join(GinkgoT().TempDir(), "zone..edge")
		Expect(os.Mkdir(dottedDir, 0o700)).To(Succeed())
		files.dotted = writeToken(filepath.Join(dottedDir, "token"), "dotted-token\n")
	})

	DescribeTable("should load the token",
		func(config func(tokenFiles) multizone.KDSClientAuthConfig, token string) {
			Expect(config(files).LoadToken()).To(Equal(token))
		},
		Entry("no token", func(tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{}
		}, ""),
		Entry("inline", func(tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenInline: " inline-token\n"}
		}, "inline-token"),
		Entry("path", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: f.token}
		}, "file-token"),
		Entry("path over inline", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenInline: "inline-token", TokenPath: f.token}
		}, "file-token"),
		Entry("dots in a file name", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: f.dotted}
		}, "dotted-token"),
		Entry("traversal resolving back in", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: f.traversing}
		}, "file-token"),
	)

	DescribeTable("should fail to load the token",
		func(config func(tokenFiles) multizone.KDSClientAuthConfig, errSubstring string) {
			_, err := config(files).LoadToken()

			Expect(err).To(MatchError(ContainSubstring(errSubstring)))
		},
		Entry("empty file", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: f.empty}
		}, "is empty"),
		Entry("blank inline", func(tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenInline: "   "}
		}, ".TokenInline is empty"),
		Entry("missing file", func(f tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: filepath.Join(filepath.Dir(f.token), "missing")}
		}, "could not read zone token"),
		Entry("escaping path", func(tokenFiles) multizone.KDSClientAuthConfig {
			return multizone.KDSClientAuthConfig{TokenPath: filepath.Join("..", "..", "etc", "token")}
		}, "traversal sequence"),
	)

	It("should sanitize the inline token", func() {
		cfg := multizone.KdsClientConfig{Auth: multizone.KDSClientAuthConfig{TokenInline: "token"}}

		cfg.Sanitize()

		Expect(cfg.Auth.TokenInline).To(Equal("*****"))
	})
})
