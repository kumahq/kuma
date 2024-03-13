import{d as m,a as s,o as i,b as g,w as t,Z as a,f as o,e,c as l,t as d,W as c,p as b}from"./index-2mpecEEN.js";import{T as u}from"./TagList-JmFVLj28.js";const w={key:0,class:"stack-with-borders"},f={key:1,class:"stack-with-borders"},C=m({__name:"ConnectionInboundSummaryOverviewView",props:{inbound:{},gateway:{}},setup(p){const n=p;return(v,x)=>{const r=s("KBadge"),_=s("AppView"),y=s("RouteView");return i(),g(y,{name:"connection-inbound-summary-overview-view"},{default:t(({t:h})=>[e(_,null,{default:t(()=>[n.gateway?(i(),l("div",w,[e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Tags
          `)]),body:t(()=>[e(u,{tags:n.gateway.tags,alignment:"right"},null,8,["tags"])]),_:1})])):n.inbound?(i(),l("div",f,[e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Tags
          `)]),body:t(()=>[e(u,{tags:n.inbound.tags,alignment:"right"},null,8,["tags"])]),_:1}),o(),e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Status
          `)]),body:t(()=>[e(r,{appearance:n.inbound.health.ready?"success":"danger"},{default:t(()=>[o(d(n.inbound.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),o(),e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Protocol
          `)]),body:t(()=>[e(r,{appearance:"info"},{default:t(()=>[o(d(h(`http.api.value.${n.inbound.protocol}`)),1)]),_:2},1024)]),_:2},1024),o(),e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Address
          `)]),body:t(()=>[e(c,{text:`${n.inbound.addressPort}`},null,8,["text"])]),_:1}),o(),e(a,{layout:"horizontal"},{title:t(()=>[o(`
            Service Address
          `)]),body:t(()=>[e(c,{text:`${n.inbound.serviceAddressPort}`},null,8,["text"])]),_:1})])):b("",!0)]),_:2},1024)]),_:1})}}});export{C as default};
