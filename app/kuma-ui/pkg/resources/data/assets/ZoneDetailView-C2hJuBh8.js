import{d as V,i as c,o as l,a as k,w as t,j as o,P as T,k as e,g as r,a1 as d,t as a,A as p,C as K,l as D,a2 as L,K as N,e as y,b as u,H as S,J as x,_ as B}from"./index-DJJJbhb4.js";import{f as A}from"./kong-icons.es321-SuStQNxq.js";import{S as H}from"./StatusBadge-D5YRdgeg.js";import{_ as R}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-BE2cRkbp.js";import"./AccordionList-BjWY1tfW.js";const M=["data-testid","innerHTML"],O={"data-testid":"detail-view-details",class:"stack"},U={class:"columns"},$=["innerHTML"],E={key:0},Z=V({__name:"ZoneDetailView",props:{data:{},notifications:{default:()=>[]}},setup(z){const s=z;return(j,F)=>{const C=c("KTooltip"),b=c("DataSource"),_=c("KCard"),g=c("AppView"),w=c("RouteView");return l(),k(w,{name:"zone-cp-detail-view"},{default:t(({t:n,uri:I})=>[o(g,null,T({default:t(()=>[e(),r("div",O,[o(_,null,{default:t(()=>{var i,m;return[r("div",U,[o(d,null,{title:t(()=>[e(a(n("http.api.property.status")),1)]),body:t(()=>[o(H,{status:s.data.state},null,8,["status"])]),_:2},1024),e(),o(b,{src:I(p(K),"/control-plane/outdated/:version",{version:((m=(i=s.data.zoneInsight.version)==null?void 0:i.kumaCp)==null?void 0:m.version)??"-"})},{default:t(({data:f})=>[o(d,{class:D({version:!0,outdated:f})},{title:t(()=>[e(a(n("zone-cps.routes.item.version"))+" ",1),f===!0?(l(),k(C,{key:0,"max-width":"300"},{content:t(()=>[r("div",{innerHTML:n("zone-cps.routes.item.version_warning")},null,8,$)]),default:t(()=>[o(p(A),{color:p(L),size:p(N)},null,8,["color","size"]),e()]),_:2},1024)):y("",!0)]),body:t(()=>{var h,v;return[e(a(((v=(h=s.data.zoneInsight.version)==null?void 0:h.kumaCp)==null?void 0:v.version)??"—"),1)]}),_:2},1032,["class"])]),_:2},1032,["src"]),e(),o(d,null,{title:t(()=>[e(a(n("http.api.property.type")),1)]),body:t(()=>[e(a(n(`common.product.environment.${s.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024),e(),o(d,null,{title:t(()=>[e(a(n("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(a(s.data.zoneInsight.authenticationType||n("common.not_applicable")),1)]),_:2},1024)])]}),_:2},1024),e(),s.data.zoneInsight.subscriptions.length>0?(l(),u("div",E,[r("h2",null,a(n("zone-cps.detail.subscriptions")),1),e(),o(_,{class:"mt-4"},{default:t(()=>[o(R,{subscriptions:s.data.zoneInsight.subscriptions},{default:t(()=>[r("p",null,a(n("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):y("",!0)])]),_:2},[s.notifications.length>0?{name:"notifications",fn:t(()=>[r("ul",null,[(l(!0),u(S,null,x(s.notifications,i=>(l(),u("li",{key:i.kind,"data-testid":`warning-${i.kind}`,innerHTML:n(`common.warnings.${i.kind}`,i.payload)},null,8,M))),128))])]),key:"0"}:void 0]),1024)]),_:1})}}}),W=B(Z,[["__scopeId","data-v-325b00cc"]]);export{W as default};
