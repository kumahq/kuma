import{d as h,a as s,o as i,b as g,w as t,W as n,f as a,e,c as l,t as d,p as b}from"./index-sL-Jj2ZJ.js";import{T as u}from"./TagList-94RPLvP4.js";import{T as c}from"./TextWithCopyButton-lp9UdEeo.js";import"./CopyButton-6B7rfdOI.js";import"./index-FZCiQto1.js";const w={key:0,class:"stack-with-borders"},f={key:1,class:"stack-with-borders"},C=h({__name:"DataPlaneInboundSummaryOverviewView",props:{inbound:{},gateway:{}},setup(p){const o=p;return(v,x)=>{const r=s("KBadge"),_=s("AppView"),m=s("RouteView");return i(),g(m,{name:"data-plane-inbound-summary-overview-view"},{default:t(({t:y})=>[e(_,null,{default:t(()=>[o.gateway?(i(),l("div",w,[e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Tags
          `)]),body:t(()=>[e(u,{tags:o.gateway.tags,alignment:"right"},null,8,["tags"])]),_:1})])):o.inbound?(i(),l("div",f,[e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Tags
          `)]),body:t(()=>[e(u,{tags:o.inbound.tags,alignment:"right"},null,8,["tags"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Status
          `)]),body:t(()=>[e(r,{appearance:o.inbound.health.ready?"success":"danger"},{default:t(()=>[a(d(o.inbound.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Protocol
          `)]),body:t(()=>[e(r,{appearance:"info"},{default:t(()=>[a(d(y(`http.api.value.${o.inbound.protocol}`)),1)]),_:2},1024)]),_:2},1024),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Address
          `)]),body:t(()=>[e(c,{text:`${o.inbound.addressPort}`},null,8,["text"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Service Address
          `)]),body:t(()=>[e(c,{text:`${o.inbound.serviceAddressPort}`},null,8,["text"])]),_:1})])):b("",!0)]),_:2},1024)]),_:1})}}});export{C as default};
