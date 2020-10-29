package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Universal In-Memory test", func() {
	RunSmokeTest(StaticConfig(`
xdsServer:
  diagnosticsPort: %d
apiServer:
  http:
    port: 0
  https:
    port: 0
dnsServer:
  port: 0
environment: universal
store:
  type: memory
`))
})
