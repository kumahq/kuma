const e="HealthCheck",s="default",t="foo-bar-baz-123",o=[{match:{service:"web"}}],a=[{match:{service:"backend"}}],c={activeChecks:{interval:"10s",timeout:"2s",unhealthyThreshold:3,healthyThreshold:4}},n={type:e,mesh:s,name:t,sources:o,destinations:a,conf:c};export{c as conf,n as default,a as destinations,s as mesh,t as name,o as sources,e as type};
