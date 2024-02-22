import{d as m,a as s,o as i,b as g,w as t,Y as n,f as a,e,c as r,t as d,V as u,p as b}from"./index-8scsg5Gp.js";import{T as c}from"./TagList-niXvlxkY.js";const w={key:0,class:"stack-with-borders"},f={key:1,class:"stack-with-borders"},z=m({__name:"DataPlaneInboundSummaryOverviewView",props:{inbound:{},gateway:{}},setup(p){const o=p;return(v,x)=>{const l=s("KBadge"),_=s("AppView"),y=s("RouteView");return i(),g(y,{name:"data-plane-inbound-summary-overview-view"},{default:t(({t:h})=>[e(_,null,{default:t(()=>[o.gateway?(i(),r("div",w,[e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Tags
          `)]),body:t(()=>[e(c,{tags:o.gateway.tags,alignment:"right"},null,8,["tags"])]),_:1})])):o.inbound?(i(),r("div",f,[e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Tags
          `)]),body:t(()=>[e(c,{tags:o.inbound.tags,alignment:"right"},null,8,["tags"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Status
          `)]),body:t(()=>[e(l,{appearance:o.inbound.health.ready?"success":"danger"},{default:t(()=>[a(d(o.inbound.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Protocol
          `)]),body:t(()=>[e(l,{appearance:"info"},{default:t(()=>[a(d(h(`http.api.value.${o.inbound.protocol}`)),1)]),_:2},1024)]),_:2},1024),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Address
          `)]),body:t(()=>[e(u,{text:`${o.inbound.addressPort}`},null,8,["text"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Service Address
          `)]),body:t(()=>[e(u,{text:`${o.inbound.serviceAddressPort}`},null,8,["text"])]),_:1})])):b("",!0)]),_:2},1024)]),_:1})}}});export{z as default};
