import{d as V,O as u,aC as g,D,aD as S,a as p,o as l,b as z,w as t,e as o,a0 as B,f as e,p as c,a1 as m,t as s,c as r,F as _,C as y,v as T}from"./index-203d56a2.js";import{S as x}from"./StatusBadge-01928c30.js";import{_ as N}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-ee2fcde5.js";import"./AccordionList-a4192120.js";const $=["data-testid","innerHTML"],A={"data-testid":"detail-view-details",class:"stack"},L={class:"columns"},Z={key:0},R=V({__name:"DetailView",props:{data:{},notifications:{default:()=>[]}},setup(h){const i=h,v=u(()=>g(i.data)),k=u(()=>D(i.data)),C=u(()=>S(i.data));return(E,F)=>{const f=p("KCard"),b=p("AppView"),w=p("RouteView");return l(),z(w,{name:"zone-cp-detail-view"},{default:t(({t:n})=>[o(b,null,B({default:t(()=>{var a;return[e(),c("div",A,[o(f,null,{body:t(()=>[c("div",L,[o(m,null,{title:t(()=>[e(s(n("http.api.property.status")),1)]),body:t(()=>[o(x,{status:k.value},null,8,["status"])]),_:2},1024),e(),o(m,null,{title:t(()=>[e(s(n("http.api.property.type")),1)]),body:t(()=>[e(s(n(`common.product.environment.${v.value||"unknown"}`)),1)]),_:2},1024),e(),o(m,null,{title:t(()=>[e(s(n("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(s(C.value||n("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),e(),(l(!0),r(_,null,y([((a=i.data.zoneInsight)==null?void 0:a.subscriptions)??[]],d=>(l(),r(_,{key:d},[d.length>0?(l(),r("div",Z,[c("h2",null,s(n("zone-cps.detail.subscriptions")),1),e(),o(f,{class:"mt-4"},{body:t(()=>[o(N,{subscriptions:d},{default:t(()=>[c("p",null,s(n("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):T("",!0)],64))),128))])]}),_:2},[i.notifications.length>0?{name:"notifications",fn:t(()=>[c("ul",null,[(l(!0),r(_,null,y(i.notifications,a=>(l(),r("li",{key:a.kind,"data-testid":`warning-${a.kind}`,innerHTML:n(`common.warnings.${a.kind}`,a.payload)},null,8,$))),128)),e()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{R as default};
