import{d as f,g as z,h as u,Y as g,aI as V,aJ as Z,o as p,l as S,j as i,w as e,n as t,H as a,k as n,a7 as _,r as l,i as h,p as c,D as b,G as x,aK as O,t as C}from"./index-bc0f9b6f.js";import{S as I}from"./StatusBadge-6c2080f6.js";import{_ as R}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-2d7e8750.js";const B={class:"stack"},T=f({__name:"ZoneSummary",props:{zoneOverview:{}},setup(r){const{t:o}=z(),s=r,d=u(()=>g(s.zoneOverview)),v=u(()=>V(s.zoneOverview)),m=u(()=>Z(s.zoneOverview));return(y,w)=>(p(),S("div",B,[i(_,null,{title:e(()=>[t(a(n(o)("http.api.property.status")),1)]),body:e(()=>[i(I,{status:d.value},null,8,["status"])]),_:1}),t(),i(_,null,{title:e(()=>[t(a(n(o)("http.api.property.type")),1)]),body:e(()=>[t(a(n(o)(`common.product.environment.${v.value||"unknown"}`)),1)]),_:1}),t(),i(_,null,{title:e(()=>[t(a(n(o)("zone-cps.routes.item.authentication_type")),1)]),body:e(()=>[t(a(m.value||n(o)("common.not_applicable")),1)]),_:1})]))}}),D=r=>(b("data-v-6806d3b4"),r=r(),x(),r),$={class:"summary-title-wrapper"},A=D(()=>c("img",{"aria-hidden":"true",src:O},null,-1)),N={class:"summary-title"},E={key:1,class:"stack"},L=f({__name:"ZoneSummaryView",props:{name:{},zoneOverview:{default:void 0}},setup(r){const{t:o}=z(),s=r;return(d,v)=>{const m=l("RouteTitle"),y=l("RouterLink"),w=l("AppView"),k=l("RouteView");return p(),h(k,{name:"zone-cp-summary-view"},{default:e(()=>[i(w,null,{title:e(()=>[c("div",$,[A,t(),c("h2",N,[i(y,{to:{name:"zone-cp-detail-view",params:{zone:s.name}}},{default:e(()=>[i(m,{title:n(o)("zone-cps.routes.item.title",{name:s.name}),render:!0},null,8,["title"])]),_:1},8,["to"])])])]),default:e(()=>[t(),s.zoneOverview===void 0?(p(),h(R,{key:0},{message:e(()=>[c("p",null,a(n(o)("common.collection.summary.empty_message",{type:"Zone"})),1)]),default:e(()=>[t(a(n(o)("common.collection.summary.empty_title",{type:"Zone"}))+" ",1)]),_:1})):(p(),S("div",E,[c("div",null,[c("h3",null,a(n(o)("zone-cps.routes.item.overview")),1),t(),i(T,{class:"mt-4","zone-overview":s.zoneOverview},null,8,["zone-overview"])])]))]),_:1})]),_:1})}}});const H=C(L,[["__scopeId","data-v-6806d3b4"]]);export{H as default};
