import{d as f,e as l,o as n,m as d,w as e,a,k as c,Y as r,b as t,t as o,S as b,p as _,a1 as m,c as p,H as g}from"./index-Bxa7bIor.js";import{_ as z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-B6OCaaNs.js";import"./AccordionList-BVgxeKWg.js";const I={class:"stack"},V={class:"columns"},v={key:0,"data-testid":"zone-ingress-subscriptions"},N=f({__name:"ZoneIngressDetailView",props:{data:{}},setup(h){const s=h;return(w,x)=>{const u=l("KCard"),k=l("AppView"),y=l("RouteView");return n(),d(y,{name:"zone-ingress-detail-view"},{default:e(({t:i})=>[a(k,null,{default:e(()=>[c("div",I,[a(u,null,{default:e(()=>[c("div",V,[a(r,null,{title:e(()=>[t(o(i("http.api.property.status")),1)]),body:e(()=>[a(b,{status:s.data.state},null,8,["status"])]),_:2},1024),t(),s.data.namespace.length>0?(n(),d(r,{key:0},{title:e(()=>[t(`
                Namespace
              `)]),body:e(()=>[t(o(s.data.namespace),1)]),_:1})):_("",!0),t(),a(r,null,{title:e(()=>[t(o(i("http.api.property.address")),1)]),body:e(()=>[s.data.zoneIngress.socketAddress.length>0?(n(),d(m,{key:0,text:s.data.zoneIngress.socketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024),t(),a(r,null,{title:e(()=>[t(o(i("http.api.property.advertisedAddress")),1)]),body:e(()=>[s.data.zoneIngress.advertisedSocketAddress.length>0?(n(),d(m,{key:0,text:s.data.zoneIngress.advertisedSocketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),t(),s.data.zoneIngressInsight.subscriptions.length>0?(n(),p("div",v,[c("h2",null,o(i("zone-ingresses.routes.item.subscriptions.title")),1),t(),a(u,{class:"mt-4"},{default:e(()=>[a(z,{subscriptions:s.data.zoneIngressInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):_("",!0)])]),_:2},1024)]),_:1})}}});export{N as default};
