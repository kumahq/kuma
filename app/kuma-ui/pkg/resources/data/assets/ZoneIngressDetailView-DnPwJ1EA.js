import{d as f,h as l,o as n,a as d,w as e,j as a,g as c,a1 as r,k as t,t as o,S as b,e as _,X as m,b as p,F as g}from"./index-ChMk9xbI.js";import{_ as z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-T-6M2sHJ.js";import"./AccordionList-DJLJWuhP.js";const I={class:"stack"},V={class:"columns"},v={key:0,"data-testid":"zone-ingress-subscriptions"},N=f({__name:"ZoneIngressDetailView",props:{data:{}},setup(h){const s=h;return(w,x)=>{const u=l("KCard"),k=l("AppView"),y=l("RouteView");return n(),d(y,{name:"zone-ingress-detail-view"},{default:e(({t:i})=>[a(k,null,{default:e(()=>[c("div",I,[a(u,null,{default:e(()=>[c("div",V,[a(r,null,{title:e(()=>[t(o(i("http.api.property.status")),1)]),body:e(()=>[a(b,{status:s.data.state},null,8,["status"])]),_:2},1024),t(),s.data.namespace.length>0?(n(),d(r,{key:0},{title:e(()=>[t(`
                Namespace
              `)]),body:e(()=>[t(o(s.data.namespace),1)]),_:1})):_("",!0),t(),a(r,null,{title:e(()=>[t(o(i("http.api.property.address")),1)]),body:e(()=>[s.data.zoneIngress.socketAddress.length>0?(n(),d(m,{key:0,text:s.data.zoneIngress.socketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024),t(),a(r,null,{title:e(()=>[t(o(i("http.api.property.advertisedAddress")),1)]),body:e(()=>[s.data.zoneIngress.advertisedSocketAddress.length>0?(n(),d(m,{key:0,text:s.data.zoneIngress.advertisedSocketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),t(),s.data.zoneIngressInsight.subscriptions.length>0?(n(),p("div",v,[c("h2",null,o(i("zone-ingresses.routes.item.subscriptions.title")),1),t(),a(u,{class:"mt-4"},{default:e(()=>[a(z,{subscriptions:s.data.zoneIngressInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):_("",!0)])]),_:2},1024)]),_:1})}}});export{N as default};
