import{d as f,r as c,o as a,i as m,w as t,j as o,p as u,a6 as g,n,H as d,k as z,a1 as w,l,F as p,I as v,m as x}from"./index-f09cca58.js";import{S as V}from"./StatusBadge-3b00ac53.js";import{_ as C}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-e271398d.js";import{g as b}from"./dataplane-0a086c06.js";import"./AccordionList-f55c6138.js";const B={class:"stack","data-testid":"detail-view-details"},A={class:"columns"},$={key:0},T=f({__name:"DetailView",props:{data:{}},setup(y){const e=y;return(N,S)=>{const _=c("KCard"),I=c("AppView"),h=c("RouteView");return a(),m(h,{name:"zone-ingress-detail-view"},{default:t(({t:r})=>[o(I,null,{default:t(()=>{var k;return[u("div",B,[o(_,null,{body:t(()=>[u("div",A,[o(g,null,{title:t(()=>[n(d(r("http.api.property.status")),1)]),body:t(()=>[o(V,{status:z(b)(e.data.zoneIngressInsight)},null,8,["status"])]),_:2},1024),n(),o(g,null,{title:t(()=>[n(d(r("http.api.property.address")),1)]),body:t(()=>{var s,i;return[(s=e.data.zoneIngress.networking)!=null&&s.address&&((i=e.data.zoneIngress.networking)!=null&&i.port)?(a(),m(w,{key:0,text:`${e.data.zoneIngress.networking.address}:${e.data.zoneIngress.networking.port}`},null,8,["text"])):(a(),l(p,{key:1},[n(d(r("common.detail.none")),1)],64))]}),_:2},1024),n(),o(g,null,{title:t(()=>[n(d(r("http.api.property.advertisedAddress")),1)]),body:t(()=>{var s,i;return[(s=e.data.zoneIngress.networking)!=null&&s.advertisedAddress&&((i=e.data.zoneIngress.networking)!=null&&i.advertisedPort)?(a(),m(w,{key:0,text:`${e.data.zoneIngress.networking.advertisedAddress}:${e.data.zoneIngress.networking.advertisedPort}`},null,8,["text"])):(a(),l(p,{key:1},[n(d(r("common.detail.none")),1)],64))]}),_:2},1024)])]),_:2},1024),n(),(a(!0),l(p,null,v([((k=e.data.zoneIngressInsight)==null?void 0:k.subscriptions)??[]],s=>(a(),l(p,{key:s},[s.length>0?(a(),l("div",$,[u("h2",null,d(r("zone-ingresses.routes.item.subscriptions.title")),1),n(),o(_,{class:"mt-4"},{body:t(()=>[o(C,{subscriptions:s},null,8,["subscriptions"])]),_:2},1024)])):x("",!0)],64))),128))])]}),_:2},1024)]),_:1})}}});export{T as default};
