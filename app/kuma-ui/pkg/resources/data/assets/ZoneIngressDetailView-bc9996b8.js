import{d as v,a as p,o as s,b as c,w as e,e as a,m,X as u,f as t,t as r,c as d,F as i,D as w,p as x}from"./index-a963f507.js";import{S as V}from"./StatusBadge-61c0813f.js";import{T as f}from"./TextWithCopyButton-442c5ee6.js";import{_ as C}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-b6e59a6f.js";import"./CopyButton-003985ad.js";import"./index-fce48c05.js";import"./AccordionList-caa11dc0.js";const z={class:"stack","data-testid":"detail-view-details"},A={class:"columns"},I={key:0},E=v({__name:"ZoneIngressDetailView",props:{data:{}},setup(h){const o=h;return(B,b)=>{const _=p("KCard"),k=p("AppView"),y=p("RouteView");return s(),c(y,{name:"zone-ingress-detail-view"},{default:e(({t:n})=>[a(k,null,{default:e(()=>{var g;return[m("div",z,[a(_,null,{default:e(()=>[m("div",A,[a(u,null,{title:e(()=>[t(r(n("http.api.property.status")),1)]),body:e(()=>[a(V,{status:o.data.state},null,8,["status"])]),_:2},1024),t(),a(u,null,{title:e(()=>[t(r(n("http.api.property.address")),1)]),body:e(()=>[o.data.zoneIngress.socketAddress.length>0?(s(),c(f,{key:0,text:o.data.zoneIngress.socketAddress},null,8,["text"])):(s(),d(i,{key:1},[t(r(n("common.detail.none")),1)],64))]),_:2},1024),t(),a(u,null,{title:e(()=>[t(r(n("http.api.property.advertisedAddress")),1)]),body:e(()=>[o.data.zoneIngress.advertisedSocketAddress.length>0?(s(),c(f,{key:0,text:o.data.zoneIngress.advertisedSocketAddress},null,8,["text"])):(s(),d(i,{key:1},[t(r(n("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),t(),(s(!0),d(i,null,w([((g=o.data.zoneIngressInsight)==null?void 0:g.subscriptions)??[]],l=>(s(),d(i,{key:l},[l.length>0?(s(),d("div",I,[m("h2",null,r(n("zone-ingresses.routes.item.subscriptions.title")),1),t(),a(_,{class:"mt-4"},{default:e(()=>[a(C,{subscriptions:l},null,8,["subscriptions"])]),_:2},1024)])):x("",!0)],64))),128))])]}),_:2},1024)]),_:1})}}});export{E as default};
