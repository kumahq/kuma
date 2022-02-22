package cmd

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Universal In-Memory test", func() {
	RunSmokeTest(StaticConfig(`
general:
  workDir: ./kuma-workdir
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
diagnostics:
  serverPort: %d
`), "./kuma-workdir")
})
