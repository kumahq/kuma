import{d as y,a as l,o as n,b as r,w as e,e as a,m as c,a5 as d,f as t,t as o,q as m,a1 as _,c as p,F as g}from"./index-UmH8j8ci.js";import{S as b}from"./StatusBadge-Cpcdk_fe.js";import{_ as z}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-CLJTp9fH.js";import"./AccordionList-B5eXKvf6.js";const I={class:"stack"},V={class:"columns"},v={key:0,"data-testid":"zone-ingress-subscriptions"},S=y({__name:"ZoneIngressDetailView",props:{data:{}},setup(h){const s=h;return(w,x)=>{const u=l("KCard"),f=l("AppView"),k=l("RouteView");return n(),r(k,{name:"zone-ingress-detail-view"},{default:e(({t:i})=>[a(f,null,{default:e(()=>[c("div",I,[a(u,null,{default:e(()=>[c("div",V,[a(d,null,{title:e(()=>[t(o(i("http.api.property.status")),1)]),body:e(()=>[a(b,{status:s.data.state},null,8,["status"])]),_:2},1024),t(),s.data.namespace.length>0?(n(),r(d,{key:0},{title:e(()=>[t(`
                Namespace
              `)]),body:e(()=>[t(o(s.data.namespace),1)]),_:1})):m("",!0),t(),a(d,null,{title:e(()=>[t(o(i("http.api.property.address")),1)]),body:e(()=>[s.data.zoneIngress.socketAddress.length>0?(n(),r(_,{key:0,text:s.data.zoneIngress.socketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024),t(),a(d,null,{title:e(()=>[t(o(i("http.api.property.advertisedAddress")),1)]),body:e(()=>[s.data.zoneIngress.advertisedSocketAddress.length>0?(n(),r(_,{key:0,text:s.data.zoneIngress.advertisedSocketAddress},null,8,["text"])):(n(),p(g,{key:1},[t(o(i("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),t(),s.data.zoneIngressInsight.subscriptions.length>0?(n(),p("div",v,[c("h2",null,o(i("zone-ingresses.routes.item.subscriptions.title")),1),t(),a(u,{class:"mt-4"},{default:e(()=>[a(z,{subscriptions:s.data.zoneIngressInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):m("",!0)])]),_:2},1024)]),_:1})}}});export{S as default};
