import{d as u,a as n,o as _,b as m,w as e,e as a,m as h,W as r,f as t,t as c}from"./index-5610cebd.js";const f={class:"stack-with-borders"},V=u({__name:"DataPlaneInboundSummaryOverviewView",props:{data:{}},setup(d){const o=d;return(w,y)=>{const s=n("KBadge"),p=n("AppView"),i=n("RouteView");return _(),m(i,{name:"data-plane-inbound-summary-overview-view"},{default:e(({t:l})=>[a(p,null,{default:e(()=>[h("div",f,[a(r,{layout:"horizontal"},{title:e(()=>[t(`
            Status
          `)]),body:e(()=>[a(s,{appearance:o.data.health.ready?"success":"danger"},{default:e(()=>[t(c(o.data.health.ready?"Healthy":"Unhealthy"),1)]),_:1},8,["appearance"])]),_:1}),t(),a(r,{layout:"horizontal"},{title:e(()=>[t(`
            Protocol
          `)]),body:e(()=>[a(s,{appearance:"info"},{default:e(()=>[t(c(l(`http.api.value.${o.data.protocol}`)),1)]),_:2},1024)]),_:2},1024)])]),_:2},1024)]),_:1})}}});export{V as default};
