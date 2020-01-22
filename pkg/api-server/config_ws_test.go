package api_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	api_server_config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config WS", func() {

	It("should return the config", func() {
		// given
		cfg := api_server_config.DefaultApiServerConfig()

		// setup
		resourceStore := memory.NewStore()
		apiServer := createTestApiServer(resourceStore, cfg)

		stop := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			err := apiServer.Start(stop)
			Expect(err).ToNot(HaveOccurred())
		}()
		port := strings.Split(apiServer.Address(), ":")[1]

		// wait for the server
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://localhost:%s/config", port))
			return err
		}, "3s").ShouldNot(HaveOccurred())

		// when
		resp, err := http.Get(fmt.Sprintf("http://localhost:%s/config", port))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		// when
		json := fmt.Sprintf(`
        {
          "apiServer": {
            "corsAllowedDomains": [
              ".*"
            ],
            "port": %s,
            "readOnly": false
          },
          "bootstrapServer": {
            "params": {
              "adminAccessLogPath": "/dev/null",
              "adminAddress": "127.0.0.1",
              "adminPort": 0,
              "xdsConnectTimeout": "1s",
              "xdsHost": "",
              "xdsPort": 0
            },
            "port": 5682
          },
          "dataplaneTokenServer": {
            "enabled": true,
            "local": {
              "port": 5679
            },
            "public": {
              "clientCertsDir": "",
              "enabled": false,
              "interface": "",
              "port": 0,
              "tlsCertFile": "",
              "tlsKeyFile": ""
            }
          },
          "adminServer": {
            "local": {
              "port": 5679
            },
            "public": {
              "clientCertsDir": "",
              "enabled": false,
              "interface": "",
              "port": 0,
              "tlsCertFile": "",
              "tlsKeyFile": ""
            },
            "apis": {
              "dataplaneToken": {
                "enabled": true
              }
            }
          },
          "defaults": {
            "mesh": "type: Mesh\nname: default\nmtls:\n  ca: {}\n  enabled: false\n"
          },
          "environment": "universal",
          "general": {
            "advertisedHostname": "localhost"
          },
          "guiServer": {
            "port": 5683,
            "apiServerUrl": ""
          },
          "monitoringAssignmentServer": {
            "assignmentRefreshInterval": "1s",
            "grpcPort": 5676
          },
          "reports": {
            "enabled": true
          },
          "runtime": {
            "kubernetes": {
              "admissionServer": {
                "address": "",
                "certDir": "",
                "port": 5443
              }
            }
          },
          "sdsServer": {
            "grpcPort": 5677,
            "tlsCertFile": "",
            "tlsKeyFile": ""
          },
          "store": {
            "kubernetes": {
              "systemNamespace": "kuma-system"
            },
            "postgres": {
              "connectionTimeout": 5,
              "dbName": "kuma",
              "host": "127.0.0.1",
              "maxOpenConnections": 0,
              "password": "*****",
              "port": 15432,
              "tls": {
                "certPath": "",
                "keyPath": "",
                "mode": "disable",
                "caPath": ""
              },
              "user": "kuma"
            },
            "type": "memory"
          },
          "xdsServer": {
            "dataplaneConfigurationRefreshInterval": "1s",
            "dataplaneStatusFlushInterval": "1s",
            "diagnosticsPort": 5680,
            "grpcPort": 5678
          }
        }
		`, port)
		Expect(body).To(MatchJSON(json))
	})
})
