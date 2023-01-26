const n=1,e=[{type:"ProxyTemplate",mesh:"hello-world",name:"pt-123",selectors:[{match:{service:"backend"}}],conf:{imports:["default-proxy"],resources:[{name:"raw-name",version:"raw-version",resource:`'@type': type.googleapis.com/envoy.api.v2.Cluster
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
`}]}}],t=null,o={total:1,items:e,next:t};export{o as default,e as items,t as next,n as total};
