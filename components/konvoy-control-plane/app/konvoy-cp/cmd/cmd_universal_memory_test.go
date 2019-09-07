package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Universal In-Memory test", func() {
	RunSmokeTest(`
xdsServer:
  grpcPort: 0
  httpPort: 0
  diagnosticsPort: %d
bootstrapServer:
  port: 0
apiServer:
  port: 0
sdsServer:
  grpcPort: 0
environment: universal
store:
  type: memory
`)
})
