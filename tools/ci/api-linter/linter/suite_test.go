package linter_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestApiLinter(t *testing.T) {
	test.RunSpecs(t, "ApiLinter Suite")
}
