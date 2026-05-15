package framework

import (
	"regexp"
	"testing"
)

func TestMainAppProcessRegexMatchesPathPrefixedCommand(t *testing.T) {
	app := &UniversalApp{}
	regex := regexp.MustCompile(app.mainAppProcessRegex("kuma-cp run --config-file /kuma/kuma-cp.conf"))

	if !regex.MatchString("/usr/bin/kuma-cp run --config-file /kuma/kuma-cp.conf") {
		t.Fatal("expected regex to match path-prefixed kuma-cp command")
	}
	if !regex.MatchString("kuma-cp run --config-file /kuma/kuma-cp.conf") {
		t.Fatal("expected regex to match plain kuma-cp command")
	}
	if regex.MatchString("/usr/bin/not-kuma-cp run --config-file /kuma/kuma-cp.conf") {
		t.Fatal("expected regex to reject a different executable")
	}
}
