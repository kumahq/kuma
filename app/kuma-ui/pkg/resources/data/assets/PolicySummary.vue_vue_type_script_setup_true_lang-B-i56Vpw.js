import{d as w,k as R,l as L,r as y,m as l,o as a,w as s,a as z,e as t,c as m,q as g,b as f,p,t as i,s as r,L as c,Y as N,F as V}from"./index-D_WxlpfD.js";const A={class:"mt-4 stack-with-borders","data-testid":"structured-view"},S={key:0},D={key:1},F={class:"mt-4"},Y=w({__name:"PolicySummary",props:{policy:{},format:{},legacy:{type:Boolean}},setup(B){const{t:d}=R(),C=L(),n=B;return(e,o)=>{const u=y("XBadge"),X=y("XAction"),b=y("XCodeBlock"),v=y("XLayout");return a(),l(v,{type:"stack"},{default:s(()=>[z(e.$slots,"header"),o[11]||(o[11]=t()),n.format==="structured"?(a(),m(V,{key:0},[g("div",A,[f(c,{layout:"horizontal"},{title:s(()=>[t(i(r(d)("http.api.property.type")),1)]),body:s(()=>[n.policy.type?(a(),l(u,{key:0,appearance:"neutral"},{default:s(()=>[t(i(n.policy.type),1)]),_:1})):p("",!0)]),_:1}),o[7]||(o[7]=t()),n.legacy?p("",!0):(a(),l(c,{key:0,layout:"horizontal"},{title:s(()=>[t(i(r(d)("http.api.property.targetRef")),1)]),body:s(()=>{var k;return[(k=n.policy.spec)!=null&&k.targetRef?(a(),l(u,{key:0,appearance:"neutral"},{default:s(()=>[t(i(n.policy.spec.targetRef.kind),1),n.policy.spec.targetRef.name?(a(),m("span",S,[o[1]||(o[1]=t(":")),g("b",null,i(n.policy.spec.targetRef.name),1)])):p("",!0)]),_:1})):(a(),l(u,{key:1,appearance:"neutral"},{default:s(()=>o[2]||(o[2]=[t(`
              Mesh
            `)])),_:1}))]}),_:1})),o[8]||(o[8]=t()),n.policy.namespace.length>0?(a(),l(c,{key:1,layout:"horizontal"},{title:s(()=>[t(i(r(d)("data-planes.routes.item.namespace")),1)]),body:s(()=>[t(i(n.policy.namespace),1)]),_:1})):p("",!0),o[9]||(o[9]=t()),r(C)("use zones")&&n.policy.zone?(a(),l(c,{key:2,layout:"horizontal"},{title:s(()=>o[5]||(o[5]=[t(`
            Zone
          `)])),body:s(()=>[f(X,{to:{name:"zone-cp-detail-view",params:{zone:n.policy.zone}}},{default:s(()=>[t(i(n.policy.zone),1)]),_:1},8,["to"])]),_:1})):p("",!0)]),o[10]||(o[10]=t()),f(b,{language:"yaml",code:r(N).stringify(e.policy.spec?{spec:e.policy.spec}:{..."sources"in e.policy?{sources:e.policy.sources}:{},..."destinations"in e.policy?{destinations:e.policy.destinations}:{},..."selectors"in e.policy?{selectors:e.policy.selectors}:{},..."conf"in e.policy?{conf:e.policy.conf}:{},..."routing"in e.policy?{routing:e.policy.routing}:{},..."tracing"in e.policy?{tracing:e.policy.tracing}:{},..."metrics"in e.policy?{metrics:e.policy.metrics}:{},..."logging"in e.policy?{logging:e.policy.logging}:{}})},null,8,["code"])],64)):(a(),m("div",D,[g("div",F,[z(e.$slots,"default")])]))]),_:3})}}});export{Y as _};
