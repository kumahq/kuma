import{g as y,A as k,D as g,S as w,T as b,_ as z}from"./RouteView.vue_vue_type_script_setup_true_lang-6a5fe479.js";import{_ as E}from"./SubscriptionList.vue_vue_type_script_setup_true_lang-8d9146b2.js";import{g as C}from"./dataplane-30467516.js";import{d as v,r as B,o as i,a as _,w as t,h as e,i as l,g as s,t as d,b as n,e as h,F as I,f as V}from"./index-109d614e.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-f6b38182.js";import"./AccordionList-c1bef8b2.js";const x={class:"stack","data-testid":"detail-view-details"},D={class:"columns",style:{"--columns":"2"}},N={key:0},Z=v({__name:"ZoneEgressDetailView",props:{data:{}},setup(f){const a=f,{t:r}=y();return(p,S)=>{const c=B("KCard");return i(),_(z,{name:"zone-egress-detail-view","data-testid":"zone-egress-detail-view"},{default:t(()=>[e(k,null,{default:t(()=>{var u;return[l("div",x,[e(c,null,{body:t(()=>[l("div",D,[e(g,null,{title:t(()=>[s(d(n(r)("http.api.property.status")),1)]),body:t(()=>[e(w,{status:n(C)(a.data.zoneEgressInsight)},null,8,["status"])]),_:1}),s(),e(g,null,{title:t(()=>[s(d(n(r)("http.api.property.address")),1)]),body:t(()=>{var o,m;return[(o=a.data.zoneEgress.networking)!=null&&o.address&&((m=a.data.zoneEgress.networking)!=null&&m.port)?(i(),_(b,{key:0,text:`${a.data.zoneEgress.networking.address}:${a.data.zoneEgress.networking.port}`},null,8,["text"])):(i(),h(I,{key:1},[s(d(n(r)("common.detail.none")),1)],64))]}),_:1})])]),_:1}),s(),(((u=p.data.zoneEgressInsight)==null?void 0:u.subscriptions)??[]).length>0?(i(),h("div",N,[l("h2",null,d(n(r)("zone-egresses.detail.subscriptions")),1),s(),e(c,{class:"mt-4"},{body:t(()=>{var o;return[e(E,{subscriptions:((o=p.data.zoneEgressInsight)==null?void 0:o.subscriptions)??[]},null,8,["subscriptions"])]}),_:1})])):V("",!0)])]}),_:1})]),_:1})}}});export{Z as default};
