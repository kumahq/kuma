// +build release

package release_test

import (
	"testing"

	"github.com/kumahq/kuma/pkg/test"
)

func TestRelease(t *testing.T) {
	test.RunSpecs(t, "Release Suite")
}
