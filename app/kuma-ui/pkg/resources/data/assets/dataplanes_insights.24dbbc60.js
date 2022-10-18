const n=10,e=[{type:"DataplaneOverview",mesh:"default",name:"backend",creationTime:"2021-02-17T08:33:36.442044+01:00",modificationTime:"2021-02-17T08:33:36.442044+01:00",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:7776,servicePort:7777,serviceAddress:"127.0.0.1",tags:{"kuma.io/protocol":"http","kuma.io/service":"backend"}}],outbound:[{port:10001,tags:{"kuma.io/service":"frontend"}}]}},dataplaneInsight:{subscriptions:[{id:"118b4d6f-7a98-4172-96d9-85ffb8b20b16",controlPlaneInstanceId:"foo",connectTime:"2021-02-17T07:33:36.412683Z",disconnectTime:"2021-02-17T07:33:36.412683Z",status:{lastUpdateTime:"2021-02-17T10:48:03.638434Z",total:{responsesSent:"5",responsesAcknowledged:"5"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"2",responsesAcknowledged:"2"},lds:{responsesSent:"2",responsesAcknowledged:"2"},rds:{}},version:{kumaDp:{version:"1.0.7",gitTag:"unknown",gitCommit:"unknown",buildDate:"unknown"},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}},{id:"118b4d6f-7a98-4172-96d9-85ffb8b20b16",controlPlaneInstanceId:"foo",connectTime:"2021-02-17T07:33:36.412683Z",status:{lastUpdateTime:"2021-02-17T10:48:03.638434Z",total:{responsesSent:"5",responsesAcknowledged:"5"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"2",responsesAcknowledged:"2"},lds:{responsesSent:"2",responsesAcknowledged:"2"},rds:{}},version:{kumaDp:{version:"1.0.7",gitTag:"unknown",gitCommit:"unknown",buildDate:"unknown",kumaCpCompatible:!0},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL",kumaDpCompatible:!0},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"frontend",creationTime:"2021-02-17T11:48:01.997694+01:00",modificationTime:"2021-02-17T11:48:01.997694+01:00",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:8887,servicePort:8888,serviceAddress:"127.0.0.1",tags:{"kuma.io/protocol":"http","kuma.io/service":"frontend"}}],outbound:[{port:10002,tags:{"kuma.io/service":"backend"}}]}},dataplaneInsight:{subscriptions:[{id:"25875983-ba09-47f2-91c4-5c0471a954ce",controlPlaneInstanceId:"foo",connectTime:"2021-02-17T10:48:01.962002Z",status:{lastUpdateTime:"2021-02-17T10:48:04.004243Z",total:{responsesSent:"3",responsesAcknowledged:"3"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"1",responsesAcknowledged:"1"},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}},version:{kumaDp:{version:"1.0.6",gitTag:"unknown",gitCommit:"unknown",buildDate:"unknown",kumaCpCompatible:!1},envoy:{version:"1.15.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL",kumaDpCompatible:!1},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"db",creationTime:"2021-02-17T11:48:01.997694+01:00",modificationTime:"2021-02-17T11:48:01.997694+01:00",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:8887,servicePort:8888,serviceAddress:"127.0.0.1",tags:{"kuma.io/protocol":"http","kuma.io/service":"db"}}],outbound:[{port:10002,tags:{"kuma.io/service":"backend"}}]}},dataplaneInsight:{subscriptions:[{id:"25875983-ba09-47f2-91c4-5c0471a954ce",controlPlaneInstanceId:"foo",connectTime:"2021-02-17T10:48:01.962002Z",status:{lastUpdateTime:"2021-02-17T10:48:04.004243Z",total:{responsesSent:"3",responsesAcknowledged:"3"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"1",responsesAcknowledged:"1"},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}},version:{kumaDp:{version:"unknown",gitTag:"unknown",gitCommit:"unknown",buildDate:"unknown"},envoy:{version:"1.15.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"no-subscriptions",creationTime:"2021-02-17T08:33:36.442044+01:00",modificationTime:"2021-02-17T08:33:36.442044+01:00",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:7776,servicePort:7777,serviceAddress:"127.0.0.1",tags:{"kuma.io/protocol":"http","kuma.io/service":"no-subscriptions"}}],outbound:[{port:10001,tags:{"kuma.io/service":"frontend"}}]}},dataplaneInsight:{subscriptions:[]}},{type:"DataplaneOverview",mesh:"default",name:"cluster-1.backend-02",creationTime:"2021-02-19T06:39:07.741335Z",modificationTime:"2021-02-19T06:39:07.741335Z",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:20012,servicePort:20013,tags:{"kuma.io/service":"backend","kuma.io/zone":"cluster-1",version:"2"}}]}},dataplaneInsight:{subscriptions:[{id:"4452a6c0-53df-4259-aead-4b0c92dc1e66",controlPlaneInstanceId:"foobar",connectTime:"2021-02-19T06:39:08.865739Z",status:{lastUpdateTime:"2021-02-19T06:39:10.355507Z",total:{responsesSent:"2",responsesAcknowledged:"2"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}},version:{kumaDp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:28:29Z"},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"cluster-1.backend-03",creationTime:"2021-02-19T06:39:20.921047Z",modificationTime:"2021-02-19T06:39:20.921047Z",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:20010,servicePort:20011,tags:{"kuma.io/service":"backend","kuma.io/zone":"cluster-1",version:"1"}}]}},dataplaneInsight:{subscriptions:[{id:"e87fbc54-936c-44d9-bc82-804f18bfb75b",controlPlaneInstanceId:"foobar",connectTime:"2021-02-19T06:39:22.073246Z",status:{lastUpdateTime:"2021-02-19T06:39:23.597417Z",total:{responsesSent:"2",responsesAcknowledged:"2"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}},version:{kumaDp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:28:29Z"},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"cluster-1.gateway-01",creationTime:"2021-02-19T06:59:59.223509Z",modificationTime:"2021-02-19T06:59:59.223509Z",dataplane:{networking:{address:"127.0.0.1",gateway:{tags:{"kuma.io/service":"gateway","kuma.io/zone":"cluster-1"},type:"BUILTIN"},outbound:[{port:10005,tags:{"kuma.io/service":"backend"}}]}},dataplaneInsight:{subscriptions:[{id:"aab878e6-fc21-4077-a077-08890fae25e2",controlPlaneInstanceId:"foobar",connectTime:"2021-02-19T07:00:00.369299Z",status:{lastUpdateTime:"2021-02-19T07:00:27.410021Z",total:{responsesSent:"4",responsesAcknowledged:"4"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"2",responsesAcknowledged:"2"},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}},version:{kumaDp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:28:29Z"},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"cluster-1.ingress-02",creationTime:"2021-02-19T06:38:49.548705Z",modificationTime:"2021-02-19T07:39:23.218956+01:00",dataplane:{networking:{address:"127.0.0.1",inbound:[{port:2e4,tags:{"kuma.io/service":"ingress","kuma.io/zone":"cluster-1"}}]}},dataplaneInsight:{subscriptions:[{id:"6faec878-cf27-42c4-804a-3c50b4021665",controlPlaneInstanceId:"foobar",connectTime:"2021-02-19T06:38:50.697460Z",status:{lastUpdateTime:"2021-02-19T06:39:22.733623Z",total:{responsesSent:"6",responsesAcknowledged:"6"},cds:{responsesSent:"2",responsesAcknowledged:"2"},eds:{responsesSent:"2",responsesAcknowledged:"2"},lds:{responsesSent:"2",responsesAcknowledged:"2"},rds:{}},version:{kumaDp:{version:"1.0.0-rc2-211-g823fe8ce",gitTag:"1.0.0-rc2-211-g823fe8ce",gitCommit:"823fe8cef6430a8f75e72a7224eb5a8ab571ec42",buildDate:"2021-02-18T13:28:29Z"},envoy:{version:"1.16.2",build:"e98e41a8e168af7acae8079fc0cd68155f699aa3/1.16.2/Modified/DEBUG/BoringSSL"},dependencies:{coredns:"1.8.3"}}}]}},{type:"DataplaneOverview",mesh:"default",name:"dataplane-test-456",creationTime:"2020-06-29T09:27:46.05334-04:00",modificationTime:"2020-06-29T09:27:46.05334-04:00",dataplane:{networking:{address:"192.168.64.8",inbound:[{port:10001,servicePort:9e3,tags:{env:"dev","kuma.io/service":"kuma-example-backend",tag01:"value01",reallyLongTagLabelHere:"a-really-long-tag-value-here"}}]}},dataplaneInsight:{subscriptions:[{id:"426fe0d8-f667-11e9-b081-acde48001122",controlPlaneInstanceId:"06070748-f667-11e9-b081-acde48001122",connectTime:"2019-10-24T14:04:56.820350Z",status:{lastUpdateTime:"2019-10-24T14:04:57.832482Z",total:{responsesSent:"3",responsesAcknowledged:"3"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"1",responsesAcknowledged:"1"},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}}}],mTLS:{certificateExpirationTime:"2021-08-20T09:45:51Z",lastCertificateRegeneration:"2021-08-20T08:45:51.869135Z",certificateRegenerations:4,issuedBackend:"ca-2",supportedBackends:["ca-2","ca-1"]}}},{type:"DataplaneOverview",mesh:"default",name:"ingress-dp-test-123",dataplane:{networking:{address:"10.0.0.1",inbound:[{port:1e4,servicePort:9e3,tags:{env:"dev","kuma.io/service":"kuma-example-backend",tag01:"value01",reallyLongTagLabelHere:"a-really-long-tag-value-here"}}]}},dataplaneInsight:{subscriptions:[{id:"426fe0d8-f667-11e9-b081-acde48001122",controlPlaneInstanceId:"06070748-f667-11e9-b081-acde48001122",connectTime:"2019-10-24T14:04:56.820350Z",status:{lastUpdateTime:"2019-10-24T14:04:57.832482Z",total:{responsesSent:"3",responsesAcknowledged:"3"},cds:{responsesSent:"1",responsesAcknowledged:"1"},eds:{responsesSent:"1",responsesAcknowledged:"1"},lds:{responsesSent:"1",responsesAcknowledged:"1"},rds:{}}}],mTLS:{certificateExpirationTime:"2020-05-11T16:53:55Z",lastCertificateRegeneration:"2020-05-11T16:53:40.862241Z",certificateRegenerations:2,issuedBackend:"ca-2",supportedBackends:["ca-2","ca-1"]}}}],s="http://localhost:5681/meshes/default/dataplanes+insights?offset=24&size=24",a={total:10,items:e,next:s};export{a as default,e as items,s as next,n as total};
