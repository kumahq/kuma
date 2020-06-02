package api_server_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api_server_config "github.com/Kong/kuma/pkg/config/api-server"
	"github.com/Kong/kuma/pkg/plugins/resources/memory"
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
            "catalog": {
              "bootstrap": {
                "url": ""
              },
              "monitoringAssignment": {
                "url": ""
              },
              "sds": {
                "url": ""
              }
            },
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
            "mesh": "type: Mesh\nname: default\n"
          },
          "environment": "universal",
          "general": {
            "advertisedHostname": "localhost"
          },
          "guiServer": {
            "port": 5683,
            "apiServerUrl": ""
          },
          "metrics": {
            "dataplane": {
              "enabled": true,
              "subscriptionLimit": 10
            }
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
              },
              "injector": {
                "cniEnabled": false,
                "initContainer": {
                  "image": "kuma/kuma-init:latest"
                },
                "sidecarContainer": {
                  "adminPort": 9901,
                  "drainTime": "30s",
                  "gid": 5678,
                  "image": "kuma/kuma-dp:latest",
                  "livenessProbe": {
                    "failureThreshold": 12,
                    "initialDelaySeconds": 60,
                    "periodSeconds": 5,
                    "timeoutSeconds": 3
                  },
                  "readinessProbe": {
                    "failureThreshold": 12,
                    "initialDelaySeconds": 1,
                    "periodSeconds": 5,
                    "successThreshold": 1,
                    "timeoutSeconds": 3
                  },
                  "redirectPort": 15001,
                  "resources": {
                    "limits": {
                      "cpu": "1000m",
                      "memory": "512Mi"
                    },
                    "requests": {
                      "cpu": "50m",
                      "memory": "64Mi"
                    }
                  },
                  "uid": 5678
                }
              }
            }
          },
          "sdsServer": {
            "grpcPort": 5677,
            "tlsCertFile": "",
            "tlsKeyFile": "",
            "dataplaneConfigurationRefreshInterval": "1s"
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
            "cache": {
              "enabled": true,
              "expirationTime": "1s"
            },
            "type": "memory"
          },
          "xdsServer": {
            "dataplaneConfigurationRefreshInterval": "1s",
            "dataplaneStatusFlushInterval": "1s",
            "diagnosticsPort": 5680,
            "grpcPort": 5678,
            "tlsCertFile": "",
            "tlsKeyFile": ""
          }
        }
		`, port)
		Expect(body).To(MatchJSON(json))
	})
})
