import{d,k as v,o as c,c as _,e as a,w as e,f as o,t as m,l as r,X as u,b as p,F as E,a as l,m as i,v as k,x,aC as S,_ as V}from"./index-0b4678e0.js";import{S as O}from"./StatusBadge-bd1b63e1.js";import{T as b}from"./TextWithCopyButton-bc8c6ef3.js";import{_ as B}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-3b71f9a1.js";import"./CopyButton-18f43ddc.js";import"./index-fce48c05.js";const C={class:"stack"},R=d({__name:"ZoneEgressSummary",props:{zoneEgressOverview:{}},setup(n){const{t:s}=v(),t=n;return(g,y)=>(c(),_("div",C,[a(u,null,{title:e(()=>[o(m(r(s)("http.api.property.status")),1)]),body:e(()=>[a(O,{status:t.zoneEgressOverview.state},null,8,["status"])]),_:1}),o(),a(u,null,{title:e(()=>[o(m(r(s)("http.api.property.address")),1)]),body:e(()=>[t.zoneEgressOverview.zoneEgress.socketAddress.length>0?(c(),p(b,{key:0,text:t.zoneEgressOverview.zoneEgress.socketAddress},null,8,["text"])):(c(),_(E,{key:1},[o(m(r(s)("common.detail.none")),1)],64))]),_:1})]))}}),Z=n=>(k("data-v-b8c65a14"),n=n(),x(),n),I={class:"summary-title-wrapper"},T=Z(()=>i("img",{"aria-hidden":"true",src:S},null,-1)),A={class:"summary-title"},N={key:1,class:"stack"},$=d({__name:"ZoneEgressSummaryView",props:{name:{},zoneEgressOverview:{default:void 0}},setup(n){const{t:s}=v(),t=n;return(g,y)=>{const w=l("RouteTitle"),f=l("RouterLink"),h=l("AppView"),z=l("RouteView");return c(),p(z,{name:"zone-egress-summary-view"},{default:e(()=>[a(h,null,{title:e(()=>[i("div",I,[T,o(),i("h2",A,[a(f,{to:{name:"zone-egress-detail-view",params:{zone:t.name}}},{default:e(()=>[a(w,{title:r(s)("zone-egresses.routes.item.title",{name:t.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:e(()=>[o(),t.zoneEgressOverview===void 0?(c(),p(B,{key:0},{message:e(()=>[i("p",null,m(r(s)("common.collection.summary.empty_message",{type:"ZoneEgress"})),1)]),default:e(()=>[o(m(r(s)("common.collection.summary.empty_title",{type:"ZoneEgress"}))+" ",1)]),_:1})):(c(),_("div",N,[i("div",null,[i("h3",null,m(r(s)("zone-egresses.routes.item.overview")),1),o(),a(R,{class:"mt-4","zone-egress-overview":t.zoneEgressOverview},null,8,["zone-egress-overview"])])]))]),_:1})]),_:1})}}});const q=V($,[["__scopeId","data-v-b8c65a14"]]);export{q as default};
