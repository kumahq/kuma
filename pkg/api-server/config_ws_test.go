package api_server_test

import (
	"fmt"
	"io"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	api_server "github.com/kumahq/kuma/pkg/api-server"
)

var _ = Describe("Config WS", func() {
	var stop func()
	var apiServer *api_server.ApiServer
	apiServer, stop = StartApiServer(NewTestApiServerConfigurer())
	AfterEach(func() {
		stop()
	})

	It("should return the config", func() {
		cfg := apiServer.Config()

		// when
		resp, err := http.Get(fmt.Sprintf("http://%s/config", apiServer.Address()))
		Expect(err).ToNot(HaveOccurred())

		// then
		body, err := io.ReadAll(resp.Body)
		Expect(err).ToNot(HaveOccurred())

		json := fmt.Sprintf(`
		{
		  "apiServer": {
			"auth": {
			  "clientCertsDir": "../../test/certs/client"
			},
			"authn": {
			  "localhostIsAdmin": true,
			  "type": "tokens",
			  "tokens": {
			    "bootstrapAdminToken": true
			  }
			},
			"corsAllowedDomains": [
			  ".*"
			],
			"http": {
			  "enabled": true,
			  "interface": "0.0.0.0",
			  "port": %d
			},
			"https": {
			  "enabled": true,
			  "interface": "0.0.0.0",
			  "port": %d,
			  "tlsCertFile": "../../test/certs/server-cert.pem",
			  "tlsKeyFile": "../../test/certs/server-key.pem"
			},
			"readOnly": false
		  },
		  "bootstrapServer": {
			"params": {
			  "adminAccessLogPath": "/dev/null",
			  "adminAddress": "127.0.0.1",
			  "adminPort": 9901,
			  "xdsConnectTimeout": "1s",
			  "xdsHost": "",
			  "xdsPort": 0
			}
		  },
		  "defaults": {
			"skipMeshCreation": false,
			"enableInboundPassthrough": true
		  },
		  "diagnostics": {
			"debugEndpoints": false,
			"serverPort": 5680
		  },
		  "dnsServer": {
			"CIDR": "240.0.0.0/4",
			"domain": "mesh",
			"serviceVipEnabled": true
		  },
		  "dpServer": {
			"auth": {
			  "type": "",
			  "useTokenPath": false
			},
			"hds": {
			  "checkDefaults": {
				"healthyThreshold": 1,
				"interval": "1s",
				"noTrafficInterval": "1s",
				"timeout": "2s",
				"unhealthyThreshold": 1
			  },
			  "enabled": true,
			  "interval": "5s",
			  "refreshInterval": "10s"
			},
			"port": 5678,
			"tlsCertFile": "",
			"tlsKeyFile": ""
		  },
		  "environment": "universal",
		  "general": {
			"dnsCacheTTL": "10s",
			"tlsCertFile": "",
			"tlsKeyFile": "",
			"workDir": ""
		  },
		  "guiServer": {
			"apiServerUrl": ""
		  },
		  "metrics": {
			"dataplane": {
			  "subscriptionLimit": 2,
			  "idleTimeout": "5m0s"
			},
			"mesh": {
			  "maxResyncTimeout": "20s",
			  "minResyncTimeout": "1s"
			},
			"zone": {
			  "subscriptionLimit": 10,
			  "idleTimeout": "5m0s"
			}
		  },
		  "mode": "standalone",
		  "monitoringAssignmentServer": {
			"apiVersions": [
			  "v1"
			],
			"assignmentRefreshInterval": "1s",
			"defaultFetchTimeout": "30s",
			"grpcPort": 0,
			"port": 5676
		  },
		  "multizone": {
			"global": {
			  "kds": {
				"grpcPort": 5685,
				"refreshInterval": "1s",
				"tlsCertFile": "",
				"tlsKeyFile": "",
				"zoneInsightFlushInterval": "10s",
				"maxMsgSize": 10485760
			  }
			},
			"zone": {
			  "kds": {
				"refreshInterval": "1s",
				"rootCaFile": "",
				"maxMsgSize": 10485760
			  }
			}
		  },
		  "reports": {
			"enabled": false
		  },
		  "runtime": {
			"kubernetes": {
			  "admissionServer": {
				"address": "",
				"certDir": "",
				"port": 5443
			  },
			  "controlPlaneServiceName": "kuma-control-plane",
			  "serviceAccountName": "system:serviceaccount:kuma-system:kuma-control-plane",
			  "injector": {
				"caCertFile": "",
				"builtinDNS": {
                  "enabled": true,
                  "port": 15053
                },
                "cniEnabled": false,
                "ebpf": {
                  "enabled": false,
                  "instanceIPEnvVarName": "INSTANCE_IP",
                  "bpffsPath": "/run/kuma/bpf",
                  "programsSourcePath": "/kuma/ebpf"
                },
				"exceptions": {
				  "labels": {
					"openshift.io/build.name": "*",
					"openshift.io/deployer-pod-for.name": "*"
				  }
				},
                "containerPatches": [],
				"initContainer": {
				  "image": "kuma/kuma-init:latest"
				},
				"sidecarContainer": {
				  "drainTime": "30s",
				  "envVars": {},
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
				  "redirectPortInbound": 15006,
				  "redirectPortInboundV6": 15010,
                  "redirectPortOutbound": 15001,
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
                },
                "sidecarTraffic": {
                  "excludeInboundPorts": [],
                  "excludeOutboundPorts": []
                },
                "virtualProbesEnabled": true,
                "virtualProbesPort": 9000,
                "exceptions": {
                  "labels": {
                    "openshift.io/build.name": "*",
                    "openshift.io/deployer-pod-for.name": "*"
                  }
                },
                "caCertFile": ""
              },
              "marshalingCacheExpirationTime": "5m0s"
            },
            "universal": {
              "dataplaneCleanupAge": "72h0m0s"
            }
          },
          "dpServer": {
            "port": 5678,
            "tlsCertFile": "",
            "tlsKeyFile": "",
            "auth": {
              "type": "",
			  "useTokenPath": false
            },
            "hds": {
              "checkDefaults": {
                "healthyThreshold": 1,
                "interval": "1s",
                "noTrafficInterval": "1s",
                "timeout": "2s",
                "unhealthyThreshold": 1
              },
              "enabled": true,
              "interval": "5s",
              "refreshInterval": "10s"
            }
          },
          "store": {
            "kubernetes": {
              "systemNamespace": "kuma-system"
            },
            "postgres": {
              "connectionTimeout": 5,
              "dbName": "kuma",
              "host": "127.0.0.1",
              "maxIdleConnections": 0,
              "maxOpenConnections": 0,
              "password": "*****",
              "port": 15432,
              "maxReconnectInterval": "1m0s",
              "minReconnectInterval": "10s",
              "maxIdleConnections": 50,
              "maxOpenConnections": 50,
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
            "upsert": {
              "conflictRetryBaseBackoff": "100ms",
              "conflictRetryMaxTimes": 5
            },
            "type": "memory",
            "unsafeDelete": false
          },
          "xdsServer": {
            "dataplaneConfigurationRefreshInterval": "1s",
            "dataplaneStatusFlushInterval": "10s",
            "dataplaneDeregistrationDelay": "10s",
            "nackBackoff": "5s"
          },
          "diagnostics": {
            "serverPort": 5680,
            "debugEndpoints": false
          },
          "access": {
            "type": "static",
            "static": {
              "adminResources": {
                "users": ["mesh-system:admin"],
                "groups": ["mesh-system:admin"]
              },
              "generateDpToken": {
                "users": ["mesh-system:admin"],
                "groups": ["mesh-system:admin"]
              },
              "generateUserToken": {
                "users": ["mesh-system:admin"],
                "groups": ["mesh-system:admin"]
              },
              "generateZoneToken": {
                "users": ["mesh-system:admin"],
                "groups": ["mesh-system:admin"]
              },
              "viewConfigDump": {
                "users": [ ],
                "groups": ["mesh-system:unauthenticated","mesh-system:authenticated"]
              },
              "viewStats": {
                "users": [ ],
                "groups": ["mesh-system:unauthenticated","mesh-system:authenticated"]
              },
              "viewClusters": {
                "users": [ ],
                "groups": ["mesh-system:unauthenticated","mesh-system:authenticated"]
              }
            }
          },
          "experimental": {
            "gatewayAPI": false,
            "kubeOutboundsAsVIPs": false
          }
        }
		`, cfg.HTTP.Port, cfg.HTTPS.Port)
		// when
		Expect(body).To(MatchJSON(json))
	})
})
