package linter_test

import (
	"testing"

	"github.com/kumahq/kuma/v3/pkg/test"
)

func TestApiLinter(t *testing.T) {
	test.RunSpecs(t, "ApiLinter Suite")
}
