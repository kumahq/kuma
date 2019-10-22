// +build integration

package cmd

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Standalone Postgres test", func() {

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
environment: universal
store:
  type: postgres
`))
})
