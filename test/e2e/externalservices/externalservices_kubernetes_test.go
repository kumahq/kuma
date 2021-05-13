package externalservices_test

import (
	"github.com/kumahq/kuma/test/e2e/externalservices"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("Test ExternalServices on Kubernetes", externalservices.ExternalServicesOnKubernetes)
