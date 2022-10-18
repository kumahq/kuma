const t=4,e=[{type:"ZoneOverview",name:"cluster-1",creationTime:"2021-02-19T08:06:15.380674+01:00",modificationTime:"2021-02-19T08:06:15.380674+01:00",zone:{enabled:!0},zoneInsight:{subscriptions:[{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":6681},"https":{"enabled":true,"interface":"0.0.0.0","port":6682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":6680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"cluster-1"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"b21265cf-f856-4214-ad1b-42539c4b20a9",globalInstanceId:"foobar",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{lastUpdateTime:"2021-02-19T07:06:16.384057Z",total:{responsesSent:"14",responsesAcknowledged:"14"},stat:{CircuitBreaker:{responsesSent:"1",responsesAcknowledged:"1"},Config:{responsesSent:"1",responsesAcknowledged:"1"},Dataplane:{responsesSent:"1",responsesAcknowledged:"1"},ExternalService:{responsesSent:"1",responsesAcknowledged:"1"},FaultInjection:{responsesSent:"1",responsesAcknowledged:"1"},HealthCheck:{responsesSent:"1",responsesAcknowledged:"1"},Mesh:{responsesSent:"1",responsesAcknowledged:"1"},ProxyTemplate:{responsesSent:"1",responsesAcknowledged:"1"},Retry:{responsesSent:"1",responsesAcknowledged:"1"},Secret:{responsesSent:"1",responsesAcknowledged:"1"},TrafficLog:{responsesSent:"1",responsesAcknowledged:"1"},TrafficPermission:{responsesSent:"1",responsesAcknowledged:"1"},TrafficRoute:{responsesSent:"1",responsesAcknowledged:"1"},TrafficTrace:{responsesSent:"1",responsesAcknowledged:"1"}}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":6681},"https":{"enabled":true,"interface":"0.0.0.0","port":6682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":6680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"cluster-1"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"3d3b7a11-e0f9-4f70-8cc9-2594318488d3",globalInstanceId:"MacBook-Pro-Bartlomiej.local-9e52",connectTime:"2021-02-19T07:07:15.535286Z",status:{lastUpdateTime:"2021-02-19T07:07:15.537654Z",total:{responsesSent:"14",responsesAcknowledged:"14"},stat:{CircuitBreaker:{responsesSent:"1",responsesAcknowledged:"1"},Config:{responsesSent:"1",responsesAcknowledged:"1"},Dataplane:{responsesSent:"1",responsesAcknowledged:"1"},ExternalService:{responsesSent:"1",responsesAcknowledged:"1"},FaultInjection:{responsesSent:"1",responsesAcknowledged:"1"},HealthCheck:{responsesSent:"1",responsesAcknowledged:"1"},Mesh:{responsesSent:"1",responsesAcknowledged:"1"},ProxyTemplate:{responsesSent:"1",responsesAcknowledged:"1"},Retry:{responsesSent:"1",responsesAcknowledged:"1"},Secret:{responsesSent:"1",responsesAcknowledged:"1"},TrafficLog:{responsesSent:"1",responsesAcknowledged:"1"},TrafficPermission:{responsesSent:"1",responsesAcknowledged:"1"},TrafficRoute:{responsesSent:"1",responsesAcknowledged:"1"},TrafficTrace:{responsesSent:"1",responsesAcknowledged:"1"}}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z",kumaCpGlobalCompatible:!1}}}]}},{type:"ZoneOverview",mesh:"default",name:"zone-1",creationTime:"2020-07-28T23:08:22.317322+07:00",modificationTime:"2020-07-28T23:08:22.317322+07:00",zone:{enabled:!0},zoneInsight:{subscriptions:[{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":6681},"https":{"enabled":true,"interface":"0.0.0.0","port":6682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":6680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"zone-1"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"466aa63b-70e8-4435-8bee-a7146e2cdf11",globalInstanceId:"66309679-ee95-4ea8-b17f-c715ca03bb38",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{total:{}},version:{kumaCp:{version:"1.2.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":6681},"https":{"enabled":true,"interface":"0.0.0.0","port":6682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":6680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"zone-1"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"f586f89c-2c4e-4f93-9a56-f0ea2ff010b7",globalInstanceId:"66309679-ee95-4ea8-b17f-c715ca03bb38",connectTime:"2020-07-28T16:08:24.760801Z",status:{lastUpdateTime:"2020-07-28T16:08:25.770774Z",total:{responsesSent:"11",responsesAcknowledged:"11"},stat:{CircuitBreaker:{responsesSent:"124",responsesAcknowledged:"4509369"},Dataplane:{responsesSent:"9018614",responsesAcknowledged:"13527859"},FaultInjection:{responsesSent:"18037104",responsesAcknowledged:"22546349"},HealthCheck:{responsesSent:"27055594",responsesAcknowledged:"31564839"},Mesh:{responsesSent:"36074084",responsesAcknowledged:"40583329"},ProxyTemplate:{responsesSent:"45092574",responsesAcknowledged:"49601819"},Secret:{responsesSent:"54111064",responsesAcknowledged:"58620309"},TrafficLog:{responsesSent:"63129554",responsesAcknowledged:"67638799"},TrafficPermission:{responsesSent:"72148044",responsesAcknowledged:"76657289"},TrafficRoute:{responsesSent:"81166534",responsesAcknowledged:"85675779"},TrafficTrace:{responsesSent:"90185024",responsesAcknowledged:"94694269"}}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z",kumaCpGlobalCompatible:!0}}}]}},{type:"ZoneOverview",mesh:"default",name:"zone-2",creationTime:"2018-07-17T16:05:36.995Z",modificationTime:"2019-07-17T18:08:41Z",zone:{enabled:!0},zoneInsight:{subscriptions:[{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":4681},"https":{"enabled":true,"interface":"0.0.0.0","port":4682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":4680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"zone-2"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"1",globalInstanceId:"node-001",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":4681},"https":{"enabled":true,"interface":"0.0.0.0","port":4682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":4680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"zone-2"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"2",globalInstanceId:"node-002",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{config:'{"apiServer":{"auth":{"allowFromLocalhost":true,"clientCertsDir":""},"corsAllowedDomains":[".*"],"http":{"enabled":true,"interface":"0.0.0.0","port":4681},"https":{"enabled":true,"interface":"0.0.0.0","port":4682,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"readOnly":false},"bootstrapServer":{"apiVersion":"v3","params":{"adminAccessLogPath":"/dev/null","adminAddress":"127.0.0.1","adminPort":0,"xdsConnectTimeout":"1s","xdsHost":"","xdsPort":5678}},"defaults":{"skipMeshCreation":false},"diagnostics":{"debugEndpoints":false,"serverPort":4680},"dnsServer":{"CIDR":"240.0.0.0/4","domain":"mesh","port":5653},"dpServer":{"auth":{"type":"dpToken"},"hds":{"checkDefaults":{"healthyThreshold":1,"interval":"1s","noTrafficInterval":"1s","timeout":"2s","unhealthyThreshold":1},"enabled":true,"interval":"5s","refreshInterval":"10s"},"port":5678,"tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key"},"environment":"universal","general":{"dnsCacheTTL":"10s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","workDir":"/Users/tomasz.wylezek/.kuma"},"guiServer":{"apiServerUrl":""},"metrics":{"dataplane":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":2},"mesh":{"maxResyncTimeout":"20s","minResyncTimeout":"1s"},"zone":{"enabled":true,"idleTimeout":"5m0s","subscriptionLimit":10}},"mode":"zone","monitoringAssignmentServer":{"apiVersions":["v1"],"assignmentRefreshInterval":"1s","defaultFetchTimeout":"30s","grpcPort":0,"port":5676},"multizone":{"global":{"kds":{"grpcPort":5685,"maxMsgSize":10485760,"refreshInterval":"1s","tlsCertFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.crt","tlsKeyFile":"/Users/tomasz.wylezek/.kuma/kuma-cp.key","zoneInsightFlushInterval":"10s"}},"zone":{"globalAddress":"grpcs://localhost:5685","kds":{"maxMsgSize":10485760,"refreshInterval":"1s","rootCaFile":""},"name":"zone-2"}},"reports":{"enabled":true},"runtime":{"kubernetes":{"admissionServer":{"address":"","certDir":"","port":5443},"controlPlaneServiceName":"kuma-control-plane","injector":{"builtinDNS":{"enabled":true,"port":15053},"caCertFile":"","cniEnabled":false,"exceptions":{"labels":{"openshift.io/build.name":"*","openshift.io/deployer-pod-for.name":"*"}},"initContainer":{"image":"kuma/kuma-init:latest"},"sidecarContainer":{"adminPort":9901,"drainTime":"30s","envVars":{},"gid":5678,"image":"kuma/kuma-dp:latest","livenessProbe":{"failureThreshold":12,"initialDelaySeconds":60,"periodSeconds":5,"timeoutSeconds":3},"readinessProbe":{"failureThreshold":12,"initialDelaySeconds":1,"periodSeconds":5,"successThreshold":1,"timeoutSeconds":3},"redirectPortInbound":15006,"redirectPortInboundV6":15010,"redirectPortOutbound":15001,"resources":{"limits":{"cpu":"1000m","memory":"512Mi"},"requests":{"cpu":"50m","memory":"64Mi"}},"uid":5678},"sidecarTraffic":{"excludeInboundPorts":[],"excludeOutboundPorts":[]},"virtualProbesEnabled":true,"virtualProbesPort":9000},"marshalingCacheExpirationTime":"5m0s"},"universal":{"dataplaneCleanupAge":"72h0m0s"}},"store":{"cache":{"enabled":true,"expirationTime":"1s"},"kubernetes":{"systemNamespace":"kuma-system"},"postgres":{"connectionTimeout":5,"dbName":"kuma","host":"127.0.0.1","maxIdleConnections":0,"maxOpenConnections":0,"maxReconnectInterval":"1m0s","minReconnectInterval":"10s","password":"*****","port":15432,"tls":{"caPath":"","certPath":"","keyPath":"","mode":"disable"},"user":"kuma"},"type":"memory","upsert":{"conflictRetryBaseBackoff":"100ms","conflictRetryMaxTimes":5}},"xdsServer":{"dataplaneConfigurationRefreshInterval":"1s","dataplaneStatusFlushInterval":"10s","nackBackoff":"5s"}}',id:"3",globalInstanceId:"node-003",connectTime:"2020-07-28T16:08:09.743141Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z",kumaCpGlobalCompatible:!1}}}]}},{type:"ZoneOverview",mesh:"default",name:"zone-3",creationTime:"2018-07-17T16:05:36.995Z",modificationTime:"2019-07-17T18:08:41Z",zone:{enabled:!0},zoneInsight:{subscriptions:[{id:"1",globalInstanceId:"node-001",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{id:"2",globalInstanceId:"node-002",connectTime:"2020-07-28T16:08:09.743141Z",disconnectTime:"2020-07-28T16:08:09.743194Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}},{id:"3",globalInstanceId:"node-003",connectTime:"2020-07-28T16:08:09.743141Z",status:{total:{}},version:{kumaCp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:22:30Z"}}}]}}],s=null,r={total:4,items:e,next:s};export{r as default,e as items,s as next,t as total};
