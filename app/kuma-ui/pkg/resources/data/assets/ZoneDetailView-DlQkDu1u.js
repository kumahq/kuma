import{d as V,r as c,o as l,m as y,w as t,b as n,R as T,e,k as r,Y as d,t as a,S as K,l as p,y as D,n as L,Z as N,K as S,p as k,c as u,F as x,s as B,q as R}from"./index-BRBWbknO.js";import{m as A}from"./kong-icons.es321-CF_c0uab.js";import{_ as H}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-CsEBAiX1.js";import"./AccordionList-D-ruZQld.js";const M=["data-testid","innerHTML"],O={"data-testid":"detail-view-details",class:"stack"},U={class:"columns"},Z=["innerHTML"],$={key:0},E=V({__name:"ZoneDetailView",props:{data:{},notifications:{default:()=>[]}},setup(z){const s=z;return(F,q)=>{const b=c("KTooltip"),C=c("DataSource"),_=c("KCard"),w=c("AppView"),g=c("RouteView");return l(),y(g,{name:"zone-cp-detail-view"},{default:t(({t:o,uri:I})=>[n(w,null,T({default:t(()=>[e(),r("div",O,[n(_,null,{default:t(()=>{var i,m;return[r("div",U,[n(d,null,{title:t(()=>[e(a(o("http.api.property.status")),1)]),body:t(()=>[n(K,{status:s.data.state},null,8,["status"])]),_:2},1024),e(),n(C,{src:I(p(D),"/control-plane/outdated/:version",{version:((m=(i=s.data.zoneInsight.version)==null?void 0:i.kumaCp)==null?void 0:m.version)??"-"})},{default:t(({data:f})=>[n(d,{class:L({version:!0,outdated:f})},{title:t(()=>[e(a(o("zone-cps.routes.item.version"))+" ",1),f===!0?(l(),y(b,{key:0,"max-width":"300"},{content:t(()=>[r("div",{innerHTML:o("zone-cps.routes.item.version_warning")},null,8,Z)]),default:t(()=>[n(p(A),{color:p(N),size:p(S)},null,8,["color","size"]),e()]),_:2},1024)):k("",!0)]),body:t(()=>{var h,v;return[e(a(((v=(h=s.data.zoneInsight.version)==null?void 0:h.kumaCp)==null?void 0:v.version)??"—"),1)]}),_:2},1032,["class"])]),_:2},1032,["src"]),e(),n(d,null,{title:t(()=>[e(a(o("http.api.property.type")),1)]),body:t(()=>[e(a(o(`common.product.environment.${s.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024),e(),n(d,null,{title:t(()=>[e(a(o("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(a(s.data.zoneInsight.authenticationType||o("common.not_applicable")),1)]),_:2},1024)])]}),_:2},1024),e(),s.data.zoneInsight.subscriptions.length>0?(l(),u("div",$,[r("h2",null,a(o("zone-cps.detail.subscriptions")),1),e(),n(_,{class:"mt-4"},{default:t(()=>[n(H,{subscriptions:s.data.zoneInsight.subscriptions},{default:t(()=>[r("p",null,a(o("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):k("",!0)])]),_:2},[s.notifications.length>0?{name:"notifications",fn:t(()=>[r("ul",null,[(l(!0),u(x,null,B(s.notifications,i=>(l(),u("li",{key:i.kind,"data-testid":`warning-${i.kind}`,innerHTML:o(`common.warnings.${i.kind}`,i.payload)},null,8,M))),128))])]),key:"0"}:void 0]),1024)]),_:1})}}}),P=R(E,[["__scopeId","data-v-325b00cc"]]);export{P as default};
