import{d as f,a as r,o,b as c,w as e,e as a,m as d,W as l,f as t,t as n,q as u,T as y,c as m,F as b}from"./index-C1qiy_FS.js";import{S as k}from"./StatusBadge-VZcN4L_G.js";import{_ as V}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-DWef6nQx.js";import"./AccordionList-Bjv_F91C.js";const w={class:"stack"},C={class:"columns"},x={key:0,"data-testid":"zone-egress-subscriptions"},D=f({__name:"ZoneEgressDetailView",props:{data:{}},setup(_){const s=_;return(z,B)=>{const p=r("KCard"),g=r("AppView"),h=r("RouteView");return o(),c(h,{name:"zone-egress-detail-view"},{default:e(({t:i})=>[a(g,null,{default:e(()=>[d("div",w,[a(p,null,{default:e(()=>[d("div",C,[a(l,null,{title:e(()=>[t(n(i("http.api.property.status")),1)]),body:e(()=>[a(k,{status:s.data.state},null,8,["status"])]),_:2},1024),t(),s.data.namespace.length>0?(o(),c(l,{key:0},{title:e(()=>[t(`
                Namespace
              `)]),body:e(()=>[t(n(s.data.namespace),1)]),_:1})):u("",!0),t(),a(l,null,{title:e(()=>[t(n(i("http.api.property.address")),1)]),body:e(()=>[s.data.zoneEgress.socketAddress.length>0?(o(),c(y,{key:0,text:s.data.zoneEgress.socketAddress},null,8,["text"])):(o(),m(b,{key:1},[t(n(i("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024),t(),s.data.zoneEgressInsight.subscriptions.length>0?(o(),m("div",x,[d("h2",null,n(i("zone-egresses.routes.item.subscriptions.title")),1),t(),a(p,{class:"mt-4"},{default:e(()=>[a(V,{subscriptions:s.data.zoneEgressInsight.subscriptions},null,8,["subscriptions"])]),_:1})])):u("",!0)])]),_:2},1024)]),_:1})}}});export{D as default};
