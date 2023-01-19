const t=1,e=[{type:"ProxyTemplate",mesh:"hello-world",name:"pt-123",selectors:[{match:{service:"backend"}}],conf:{imports:["default-proxy"],resources:[{name:"raw-name",version:"raw-version",resource:`'@type': type.googleapis.com/envoy.api.v2.Cluster
connectTimeout: 5s
loadAssignment:
  clusterName: localhost:8443
  endpoints:
    - lbEndpoints:
        - endpoint:
            address:
              socketAddress:
                address: 127.0.0.1
                portValue: 8443
name: localhost:8443
type: STATIC
`}]}}],n={total:1,items:e};export{n as default,e as items,t as total};
