import{d as z,r as c,o as a,i as u,w as t,j as o,p as m,a9 as g,n,H as d,W as f,k as v,a5 as w,l,F as p,I as x,m as V}from"./index-b94d59a3.js";import{_ as C}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-d3d4946e.js";import{g as b}from"./dataplane-0a086c06.js";import"./AccordionList-92754f14.js";const B={class:"stack","data-testid":"detail-view-details"},A={class:"columns"},$={key:0},R=z({__name:"DetailView",props:{data:{}},setup(y){const e=y;return(N,D)=>{const _=c("KCard"),I=c("AppView"),h=c("RouteView");return a(),u(h,{name:"zone-ingress-detail-view"},{default:t(({t:r})=>[o(I,null,{default:t(()=>{var k;return[m("div",B,[o(_,null,{body:t(()=>[m("div",A,[o(g,null,{title:t(()=>[n(d(r("http.api.property.status")),1)]),body:t(()=>[o(f,{status:v(b)(e.data.zoneIngressInsight)},null,8,["status"])]),_:2},1024),n(),o(g,null,{title:t(()=>[n(d(r("http.api.property.address")),1)]),body:t(()=>{var s,i;return[(s=e.data.zoneIngress.networking)!=null&&s.address&&((i=e.data.zoneIngress.networking)!=null&&i.port)?(a(),u(w,{key:0,text:`${e.data.zoneIngress.networking.address}:${e.data.zoneIngress.networking.port}`},null,8,["text"])):(a(),l(p,{key:1},[n(d(r("common.detail.none")),1)],64))]}),_:2},1024),n(),o(g,null,{title:t(()=>[n(d(r("http.api.property.advertisedAddress")),1)]),body:t(()=>{var s,i;return[(s=e.data.zoneIngress.networking)!=null&&s.advertisedAddress&&((i=e.data.zoneIngress.networking)!=null&&i.advertisedPort)?(a(),u(w,{key:0,text:`${e.data.zoneIngress.networking.advertisedAddress}:${e.data.zoneIngress.networking.advertisedPort}`},null,8,["text"])):(a(),l(p,{key:1},[n(d(r("common.detail.none")),1)],64))]}),_:2},1024)])]),_:2},1024),n(),(a(!0),l(p,null,x([((k=e.data.zoneIngressInsight)==null?void 0:k.subscriptions)??[]],s=>(a(),l(p,{key:s},[s.length>0?(a(),l("div",$,[m("h2",null,d(r("zone-ingresses.routes.item.subscriptions.title")),1),n(),o(_,{class:"mt-4"},{body:t(()=>[o(C,{subscriptions:s},null,8,["subscriptions"])]),_:2},1024)])):V("",!0)],64))),128))])]}),_:2},1024)]),_:1})}}});export{R as default};
