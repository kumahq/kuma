package service_identity_injection_test

import (
	. "github.com/onsi/ginkgo"

	"github.com/kumahq/kuma/test/e2e/service_identity_injection"
)

var _ = Describe("Test Service Identity Injection on Universal deployment", service_identity_injection.ServiceIdentityInjectionUniversal)
