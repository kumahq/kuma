import{d as V,f as p,aM as b,aJ as S,aN as T,r as u,o as i,g as B,w as t,h as n,a8 as D,l as e,m as r,a9 as m,C as l,S as N,j as c,F as _,D as h,k as x}from"./index-77212499.js";import{_ as Z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-a1e3e053.js";import"./AccordionList-64a7926a.js";const $=["data-testid","innerHTML"],z={"data-testid":"detail-view-details",class:"stack"},A={class:"columns",style:{"--columns":"3"}},L={key:0},P=V({__name:"ZoneDetailView",props:{data:{},notifications:{default:()=>[]}},setup(f){const s=f,v=p(()=>b(s.data)),k=p(()=>S(s.data)),C=p(()=>T(s.data));return(M,E)=>{const y=u("KCard"),g=u("AppView"),w=u("RouteView");return i(),B(w,{name:"zone-cp-detail-view"},{default:t(({t:a})=>[n(g,null,D({default:t(()=>{var o;return[e(),r("div",z,[n(y,null,{body:t(()=>[r("div",A,[n(m,null,{title:t(()=>[e(l(a("http.api.property.status")),1)]),body:t(()=>[n(N,{status:k.value},null,8,["status"])]),_:2},1024),e(),n(m,null,{title:t(()=>[e(l(a("http.api.property.type")),1)]),body:t(()=>[e(l(a(`common.product.environment.${v.value||"unknown"}`)),1)]),_:2},1024),e(),n(m,null,{title:t(()=>[e(l(a("http.api.property.authenticationType")),1)]),body:t(()=>[e(l(C.value||a("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),e(),(i(!0),c(_,null,h([((o=s.data.zoneInsight)==null?void 0:o.subscriptions)??[]],d=>(i(),c(_,{key:d},[d.length>0?(i(),c("div",L,[r("h2",null,l(a("zone-cps.detail.subscriptions")),1),e(),n(y,{class:"mt-4"},{body:t(()=>[n(Z,{subscriptions:d},null,8,["subscriptions"])]),_:2},1024)])):x("",!0)],64))),128))])]}),_:2},[s.notifications.length>0?{name:"notifications",fn:t(()=>[r("ul",null,[(i(!0),c(_,null,h(s.notifications,o=>(i(),c("li",{key:o.kind,"data-testid":`warning-${o.kind}`,innerHTML:a(`common.warnings.${o.kind}`,o.payload)},null,8,$))),128)),e()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{P as default};
