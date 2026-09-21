package multizone

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKDSServerAuthConfigValidate(t *testing.T) {
	cases := []struct {
		name      string
		authType  KDSAuthType
		errSubstr string
	}{
		{name: "none", authType: KDSAuthNone},
		{name: "zoneToken", authType: KDSAuthZoneToken},
		{name: "empty", authType: "", errSubstr: ".Type cannot be empty"},
		{name: "type a distribution registers", authType: "custom"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := KDSServerAuthConfig{Type: c.authType}.Validate()
			if c.errSubstr == "" {
				if err != nil {
					t.Fatalf("expected no validation error, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), c.errSubstr) {
				t.Fatalf("expected error to contain %q, got %v", c.errSubstr, err)
			}
		})
	}
}

func TestKDSClientAuthConfigLoadToken(t *testing.T) {
	tokenPath := filepath.Join(t.TempDir(), "token")
	if err := os.WriteFile(tokenPath, []byte("file-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tokenDir := filepath.Dir(tokenPath)
	// built by hand, filepath.Join would clean the ".." away before LoadToken sees it
	traversingPath := strings.Join([]string{tokenDir, "..", filepath.Base(tokenDir), "token"}, string(filepath.Separator))
	emptyPath := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(emptyPath, []byte("\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	dottedDir := filepath.Join(t.TempDir(), "zone..edge")
	if err := os.Mkdir(dottedDir, 0o700); err != nil {
		t.Fatal(err)
	}
	dottedPath := filepath.Join(dottedDir, "token")
	if err := os.WriteFile(dottedPath, []byte("dotted-token\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name      string
		cfg       KDSClientAuthConfig
		token     string
		errSubstr string
	}{
		{name: "no token", cfg: KDSClientAuthConfig{}},
		{name: "inline", cfg: KDSClientAuthConfig{TokenInline: " inline-token\n"}, token: "inline-token"},
		{name: "path", cfg: KDSClientAuthConfig{TokenPath: tokenPath}, token: "file-token"},
		{name: "path over inline", cfg: KDSClientAuthConfig{TokenInline: "inline-token", TokenPath: tokenPath}, token: "file-token"},
		{name: "empty file", cfg: KDSClientAuthConfig{TokenPath: emptyPath}, errSubstr: "is empty"},
		{name: "blank inline", cfg: KDSClientAuthConfig{TokenInline: "   "}, errSubstr: ".TokenInline is empty"},
		{name: "dots in a file name", cfg: KDSClientAuthConfig{TokenPath: dottedPath}, token: "dotted-token"},
		{name: "missing file", cfg: KDSClientAuthConfig{TokenPath: filepath.Join(t.TempDir(), "missing")}, errSubstr: "could not read zone token"},
		{name: "escaping path", cfg: KDSClientAuthConfig{TokenPath: filepath.Join("..", "..", "etc", "token")}, errSubstr: "traversal sequence"},
		{name: "traversal resolving back in", cfg: KDSClientAuthConfig{TokenPath: traversingPath}, token: "file-token"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			token, err := c.cfg.LoadToken()
			if c.errSubstr != "" {
				if err == nil || !strings.Contains(err.Error(), c.errSubstr) {
					t.Fatalf("expected error to contain %q, got %v", c.errSubstr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			if token != c.token {
				t.Fatalf("expected token %q, got %q", c.token, token)
			}
		})
	}
}

func TestKdsClientConfigSanitize(t *testing.T) {
	cfg := KdsClientConfig{Auth: KDSClientAuthConfig{TokenInline: "token"}}

	cfg.Sanitize()

	if cfg.Auth.TokenInline != "*****" {
		t.Fatalf("expected the token to be sanitized, got %q", cfg.Auth.TokenInline)
	}
}
