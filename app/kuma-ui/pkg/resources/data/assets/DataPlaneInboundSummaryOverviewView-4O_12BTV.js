import{d as _,a as r,o as m,b as h,w as t,X as n,f as a,e,t as i,m as y}from"./index-ANwvg_A1.js";import{T as f}from"./TagList-0w37gz68.js";import{T as l}from"./TextWithCopyButton-Ac0tj8q8.js";import"./CopyButton-XxMKSpD7.js";import"./index-FZCiQto1.js";const g={class:"stack-with-borders"},z=_({__name:"DataPlaneInboundSummaryOverviewView",props:{data:{}},setup(d){const o=d;return(w,v)=>{const s=r("KBadge"),p=r("AppView"),c=r("RouteView");return m(),h(c,{name:"data-plane-inbound-summary-overview-view"},{default:t(({t:u})=>[e(p,null,{default:t(()=>[y("div",g,[e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Tags
          `)]),body:t(()=>[e(f,{tags:o.data.tags,alignment:"right"},null,8,["tags"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Status
          `)]),body:t(()=>[e(s,{appearance:o.data.health.ready?"success":"danger"},{default:t(()=>[a(i(o.data.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Protocol
          `)]),body:t(()=>[e(s,{appearance:"info"},{default:t(()=>[a(i(u(`http.api.value.${o.data.protocol}`)),1)]),_:2},1024)]),_:2},1024),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Address
          `)]),body:t(()=>[e(l,{text:`${o.data.addressPort}`},null,8,["text"])]),_:1}),a(),e(n,{layout:"horizontal"},{title:t(()=>[a(`
            Service Address
          `)]),body:t(()=>[e(l,{text:`${o.data.serviceAddressPort}`},null,8,["text"])]),_:1})])]),_:2},1024)]),_:1})}}});export{z as default};
