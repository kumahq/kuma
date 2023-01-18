const e="ProxyTemplate",n="hello-world",s="pt-123",o=[{match:{service:"backend"}}],t={imports:["default-proxy"],resources:[{name:"raw-name",version:"raw-version",resource:`'@type': type.googleapis.com/envoy.api.v2.Cluster
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
`}]},a={type:e,mesh:n,name:s,selectors:o,conf:t};export{t as conf,a as default,n as mesh,s as name,o as selectors,e as type};
