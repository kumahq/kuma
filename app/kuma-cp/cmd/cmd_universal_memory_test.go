package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Universal In-Memory test", func() {
	RunSmokeTest(StaticConfig(`
xdsServer:
  grpcPort: 0
  diagnosticsPort: %d
bootstrapServer:
  port: 0
apiServer:
  port: 0
sdsServer:
  grpcPort: 0
dataplaneTokenServer:
  local:
    port: 0
guiServer:
  port: 0
dnsServer:
  port: 0
environment: universal
store:
  type: memory
`))
})
