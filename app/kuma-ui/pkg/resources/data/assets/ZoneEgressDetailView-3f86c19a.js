import{d as y,a as l,o as a,b as g,w as e,e as t,p,a1 as f,f as n,t as i,q as C,c as d,F as c,C as V,v as z}from"./index-784d2bbf.js";import{S as E}from"./StatusBadge-a6acfbee.js";import{T as v}from"./TextWithCopyButton-7ef74197.js";import{_ as x}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-45014a53.js";import{g as b}from"./dataplane-dcd0858b.js";import"./CopyButton-9c00109a.js";import"./index-9dd3e7d3.js";import"./AccordionList-e1625b82.js";const B={class:"stack","data-testid":"detail-view-details"},I={class:"columns"},N={key:0},W=y({__name:"ZoneEgressDetailView",props:{data:{}},setup(h){const s=h;return(S,D)=>{const m=l("KCard"),w=l("AppView"),k=l("RouteView");return a(),g(k,{name:"zone-egress-detail-view"},{default:e(({t:r})=>[t(w,null,{default:e(()=>{var u;return[p("div",B,[t(m,null,{body:e(()=>[p("div",I,[t(f,null,{title:e(()=>[n(i(r("http.api.property.status")),1)]),body:e(()=>[t(E,{status:C(b)(s.data.zoneEgressInsight)},null,8,["status"])]),_:2},1024),n(),t(f,null,{title:e(()=>[n(i(r("http.api.property.address")),1)]),body:e(()=>{var o,_;return[(o=s.data.zoneEgress.networking)!=null&&o.address&&((_=s.data.zoneEgress.networking)!=null&&_.port)?(a(),g(v,{key:0,text:`${s.data.zoneEgress.networking.address}:${s.data.zoneEgress.networking.port}`},null,8,["text"])):(a(),d(c,{key:1},[n(i(r("common.detail.none")),1)],64))]}),_:2},1024)])]),_:2},1024),n(),(a(!0),d(c,null,V([((u=s.data.zoneEgressInsight)==null?void 0:u.subscriptions)??[]],o=>(a(),d(c,{key:o},[o.length>0?(a(),d("div",N,[p("h2",null,i(r("zone-egresses.routes.item.subscriptions.title")),1),n(),t(m,{class:"mt-4"},{body:e(()=>[t(x,{subscriptions:o},null,8,["subscriptions"])]),_:2},1024)])):z("",!0)],64))),128))])]}),_:2},1024)]),_:1})}}});export{W as default};
