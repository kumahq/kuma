import{d as w,l as x,N as y,o as a,c as p,e as d,w as s,f as t,t as i,q as n,a1 as g,b as _,F as h,a as u,p as m,z as V,A as b,aE as k,_ as A}from"./index-81fc4a03.js";import{S as O}from"./StatusBadge-7904616b.js";import{T as z}from"./TextWithCopyButton-813c7852.js";import{g as $}from"./dataplane-dcd0858b.js";import{_ as B}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-10dee361.js";import"./CopyButton-9e69d8bc.js";import"./index-9dd3e7d3.js";const R={class:"stack"},Z=w({__name:"ZoneIngressSummary",props:{zoneIngressOverview:{}},setup(l){const{t:o}=x(),r=l,I=y(()=>$(r.zoneIngressOverview.zoneIngressInsight)),v=y(()=>{const{networking:e}=r.zoneIngressOverview.zoneIngress;return e!=null&&e.address&&(e!=null&&e.port)?`${e.address}:${e.port}`:null}),c=y(()=>{const{networking:e}=r.zoneIngressOverview.zoneIngress;return e!=null&&e.advertisedAddress&&(e!=null&&e.advertisedPort)?`${e.advertisedAddress}:${e.advertisedPort}`:null});return(e,f)=>(a(),p("div",R,[d(g,null,{title:s(()=>[t(i(n(o)("http.api.property.status")),1)]),body:s(()=>[d(O,{status:I.value},null,8,["status"])]),_:1}),t(),d(g,null,{title:s(()=>[t(i(n(o)("http.api.property.address")),1)]),body:s(()=>[v.value?(a(),_(z,{key:0,text:v.value},null,8,["text"])):(a(),p(h,{key:1},[t(i(n(o)("common.detail.none")),1)],64))]),_:1}),t(),d(g,null,{title:s(()=>[t(i(n(o)("http.api.property.advertisedAddress")),1)]),body:s(()=>[c.value?(a(),_(z,{key:0,text:c.value},null,8,["text"])):(a(),p(h,{key:1},[t(i(n(o)("common.detail.none")),1)],64))]),_:1})]))}}),C=l=>(V("data-v-4b74d6d8"),l=l(),b(),l),T={class:"summary-title-wrapper"},N=C(()=>m("img",{"aria-hidden":"true",src:k},null,-1)),F={class:"summary-title"},D={key:1,class:"stack"},E=w({__name:"ZoneIngressSummaryView",props:{name:{},zoneIngressOverview:{default:void 0}},setup(l){const{t:o}=x(),r=l;return(I,v)=>{const c=u("RouteTitle"),e=u("RouterLink"),f=u("AppView"),S=u("RouteView");return a(),_(S,{name:"zone-ingress-summary-view"},{default:s(()=>[d(f,null,{title:s(()=>[m("div",T,[N,t(),m("h2",F,[d(e,{to:{name:"zone-ingress-detail-view",params:{zone:r.name}}},{default:s(()=>[d(c,{title:n(o)("zone-ingresses.routes.item.title",{name:r.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:s(()=>[t(),r.zoneIngressOverview===void 0?(a(),_(B,{key:0},{message:s(()=>[m("p",null,i(n(o)("common.collection.summary.empty_message",{type:"ZoneIngress"})),1)]),default:s(()=>[t(i(n(o)("common.collection.summary.empty_title",{type:"ZoneIngress"}))+" ",1)]),_:1})):(a(),p("div",D,[m("div",null,[m("h3",null,i(n(o)("zone-ingresses.routes.item.overview")),1),t(),d(Z,{class:"mt-4","zone-ingress-overview":r.zoneIngressOverview},null,8,["zone-ingress-overview"])])]))]),_:1})]),_:1})}}});const J=A(E,[["__scopeId","data-v-4b74d6d8"]]);export{J as default};
