import{d as h,a as d,o as c,b as y,w as t,e as o,R as b,f as e,m as i,a7 as r,t as s,c as p,q as k,F as v,G as w}from"./index-BRR4OZXP.js";import{S as V}from"./StatusBadge-qnkeaeku.js";import{_ as g}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-COUaxKyD.js";import"./AccordionList-BjuAwtfg.js";const z=["data-testid","innerHTML"],C={"data-testid":"detail-view-details",class:"stack"},B={class:"columns"},x={key:0},L=h({__name:"ZoneDetailView",props:{data:{},notifications:{default:()=>[]}},setup(m){const a=m;return(I,N)=>{const u=d("KCard"),_=d("AppView"),f=d("RouteView");return c(),y(f,{name:"zone-cp-detail-view"},{default:t(({t:n})=>[o(_,null,b({default:t(()=>[e(),i("div",C,[o(u,null,{default:t(()=>[i("div",B,[o(r,null,{title:t(()=>[e(s(n("http.api.property.status")),1)]),body:t(()=>[o(V,{status:a.data.state},null,8,["status"])]),_:2},1024),e(),o(r,null,{title:t(()=>[e(s(n("http.api.property.type")),1)]),body:t(()=>[e(s(n(`common.product.environment.${a.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024),e(),o(r,null,{title:t(()=>[e(s(n("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(s(a.data.zoneInsight.authenticationType||n("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),e(),a.data.zoneInsight.subscriptions.length>0?(c(),p("div",x,[i("h2",null,s(n("zone-cps.detail.subscriptions")),1),e(),o(u,{class:"mt-4"},{default:t(()=>[o(g,{subscriptions:a.data.zoneInsight.subscriptions},{default:t(()=>[i("p",null,s(n("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):k("",!0)])]),_:2},[a.notifications.length>0?{name:"notifications",fn:t(()=>[i("ul",null,[(c(!0),p(v,null,w(a.notifications,l=>(c(),p("li",{key:l.kind,"data-testid":`warning-${l.kind}`,innerHTML:n(`common.warnings.${l.kind}`,l.payload)},null,8,z))),128))])]),key:"0"}:void 0]),1024)]),_:1})}}});export{L as default};
