const t=2,e=[{type:"ProxyTemplate",mesh:"default",name:"pt-1",selectors:[{match:{service:"backend"}}],conf:{imports:["default-proxy"],resources:[{name:"raw-name",version:"raw-version",resource:`'@type': type.googleapis.com/envoy.api.v2.Cluster
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
`}]}},{type:"ProxyTemplate",mesh:"hello-world",name:"pt-123",selectors:[{match:{service:"backend"}}],conf:{imports:["default-proxy"],resources:[{name:"raw-name",version:"raw-version",resource:`'@type': type.googleapis.com/envoy.api.v2.Cluster
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
`}]}}],n=null,s={total:2,items:e,next:n};export{s as default,e as items,n as next,t as total};
