import{d as V,r as l,o as r,m as k,w as t,b as a,l as u,z as L,T,e,k as d,Z as p,t as i,S,n as B,a3 as D,N as O,p as C,c as _,L as v,M as A,q as K}from"./index-CeTpyiyE.js";import{m as x}from"./kong-icons.es338-F9-UZ7M9.js";import{_ as E}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-ln-LXy7S.js";import"./AccordionList-DkA5-2Cm.js";const M=["data-testid","innerHTML"],R={"data-testid":"detail-view-details",class:"stack"},Z={class:"columns"},H=["innerHTML"],U={key:0},$=V({__name:"ZoneDetailView",props:{data:{}},setup(y){const n=y;return(G,P)=>{const b=l("KTooltip"),m=l("KCard"),I=l("AppView"),g=l("DataSource"),w=l("RouteView");return r(),k(w,{name:"zone-cp-detail-view"},{default:t(({t:o,uri:N})=>{var h,f;return[a(g,{src:N(u(L),"/control-plane/outdated/:version",{version:((f=(h=n.data.zoneInsight.version)==null?void 0:h.kumaCp)==null?void 0:f.version)??"-"})},{default:t(({data:s})=>[a(I,{docs:o("zones.href.docs.cta")},T({default:t(()=>[e(),d("div",R,[a(m,null,{default:t(()=>[d("div",Z,[a(p,null,{title:t(()=>[e(i(o("http.api.property.status")),1)]),body:t(()=>[a(S,{status:n.data.state},null,8,["status"])]),_:2},1024),e(),a(p,{class:B({version:!0,outdated:s==null?void 0:s.outdated})},{title:t(()=>[e(i(o("zone-cps.routes.item.version"))+" ",1),(s==null?void 0:s.outdated)===!0?(r(),k(b,{key:0,"max-width":"300"},{content:t(()=>[d("div",{innerHTML:o("zone-cps.routes.item.version_warning")},null,8,H)]),default:t(()=>[a(u(x),{color:u(D),size:u(O)},null,8,["color","size"]),e()]),_:2},1024)):C("",!0)]),body:t(()=>{var c,z;return[e(i(((z=(c=n.data.zoneInsight.version)==null?void 0:c.kumaCp)==null?void 0:z.version)??"—"),1)]}),_:2},1032,["class"]),e(),a(p,null,{title:t(()=>[e(i(o("http.api.property.type")),1)]),body:t(()=>[e(i(o(`common.product.environment.${n.data.zoneInsight.environment||"unknown"}`)),1)]),_:2},1024),e(),a(p,null,{title:t(()=>[e(i(o("zone-cps.routes.item.authentication_type")),1)]),body:t(()=>[e(i(n.data.zoneInsight.authenticationType||o("common.not_applicable")),1)]),_:2},1024)])]),_:2},1024),e(),n.data.zoneInsight.subscriptions.length>0?(r(),_("div",U,[d("h2",null,i(o("zone-cps.detail.subscriptions")),1),e(),a(m,{class:"mt-4"},{default:t(()=>[a(E,{subscriptions:n.data.zoneInsight.subscriptions},{default:t(()=>[d("p",null,i(o("zone-cps.routes.item.subscription_intro")),1)]),_:2},1032,["subscriptions"])]),_:2},1024)])):C("",!0)])]),_:2},[n.data.warnings.length>0?{name:"notifications",fn:t(()=>[d("ul",null,[(r(!0),_(v,null,A(n.data.warnings,c=>(r(),_("li",{key:c.kind,"data-testid":`warning-${c.kind}`,innerHTML:o(`common.warnings.${c.kind}`,{...c.payload,...c.kind==="INCOMPATIBLE_ZONE_AND_GLOBAL_CPS_VERSIONS"?{globalCpVersion:(s==null?void 0:s.version)??""}:{}})},null,8,M))),128))])]),key:"0"}:void 0]),1032,["docs"])]),_:2},1032,["src"])]}),_:1})}}}),Q=K($,[["__scopeId","data-v-6fd72b6d"]]);export{Q as default};
