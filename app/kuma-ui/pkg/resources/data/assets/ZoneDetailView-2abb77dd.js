import{d as v,a as p,o as c,b as w,w as t,e as a,$ as V,f as e,p as d,a0 as u,t as i,c as r,F as _,J as f,v as b}from"./index-d015481a.js";import{S as g}from"./StatusBadge-ed77f93c.js";import{_ as z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-faa926a7.js";import"./AccordionList-cad8a61c.js";const C=["data-testid","innerHTML"],B={"data-testid":"detail-view-details",class:"stack"},$={class:"columns"},x={key:0},A=v({__name:"ZoneDetailView",props:{data:{},notifications:{default:()=>[]}},setup(h){const l=h;return(N,S)=>{const m=p("KCard"),y=p("AppView"),k=p("RouteView");return c(),w(k,{name:"zone-cp-detail-view"},{default:t(({t:o})=>[a(y,null,V({default:t(()=>{var s;return[e(),d("div",B,[a(m,null,{default:t(()=>[d("div",$,[a(u,null,{title:t(()=>[e(i(o("http.api.property.status")),1)]),body:t(()=>[a(g,{status:l.data.state},null,8,["status"])]),_:2},1024),e(),a(u,null,{title:t(()=>[e(i(o("http.api.property.type")),1)]),body:t(()=>{var n;return[e(i(o(`common.product.environment.${((n=l.data.zoneInsight)==null?void 0:n.environment)||"unknown"}`)),1)]}),_:2},1024),e(),a(u,null,{title:t(()=>[e(i(o("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>{var n;return[e(i(((n=l.data.zoneInsight)==null?void 0:n.authenticationType)||o("common.not_applicable")),1)]}),_:2},1024)])]),_:2},1024),e(),(c(!0),r(_,null,f([((s=l.data.zoneInsight)==null?void 0:s.subscriptions)??[]],n=>(c(),r(_,{key:n},[n.length>0?(c(),r("div",x,[d("h2",null,i(o("zone-cps.detail.subscriptions")),1),e(),a(m,{class:"mt-4"},{default:t(()=>[a(z,{subscriptions:n},{default:t(()=>[d("p",null,i(o("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):b("",!0)],64))),128))])]}),_:2},[l.notifications.length>0?{name:"notifications",fn:t(()=>[d("ul",null,[(c(!0),r(_,null,f(l.notifications,s=>(c(),r("li",{key:s.kind,"data-testid":`warning-${s.kind}`,innerHTML:o(`common.warnings.${s.kind}`,s.payload)},null,8,C))),128)),e()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{A as default};
