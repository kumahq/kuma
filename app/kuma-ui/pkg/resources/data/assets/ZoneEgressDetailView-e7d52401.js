import{d as g,u as k,q as n,o as e,a as t,w as m,h as i,b as c,g as w,k as z,e as E}from"./index-bdbf5b57.js";import{_ as b}from"./ZoneEgressDetails.vue_vue_type_script_setup_true_lang-41207b4b.js";import{g as h,i as y,h as x,A as B,_ as V}from"./RouteView.vue_vue_type_script_setup_true_lang-9e62d24f.js";import{_ as $}from"./RouteTitle.vue_vue_type_script_setup_true_lang-dcec85af.js";import{_ as A}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-ee434af6.js";import{E as N}from"./ErrorBlock-b1fa7c54.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-32c93dbe.js";import"./kongponents.es-21ce59a5.js";import"./DefinitionListItem-310ce025.js";import"./EnvoyData-3956b32a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-5891fe23.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-dedd65ac.js";import"./TabsWidget-b5ed50b6.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-8088f9cb.js";const C={class:"zone-details"},D={key:3,class:"kcard-border","data-testid":"detail-view-details"},S=g({__name:"ZoneEgressDetailView",setup(O){const _=h(),p=k(),{t:l}=y(),s=n(null),o=n(!0),r=n(null);d();function d(){f()}async function f(){o.value=!0,r.value=null;const u=p.params.zoneEgress;try{s.value=await _.getZoneEgressOverview({name:u})}catch(a){s.value=null,a instanceof Error?r.value=a:console.error(a)}finally{o.value=!1}}return(u,a)=>(e(),t(V,null,{default:m(({route:v})=>[i($,{title:c(l)("zone-egresses.routes.item.title",{name:v.params.zoneEgress})},null,8,["title"]),w(),i(B,{breadcrumbs:[{to:{name:"zone-egress-list-view"},text:c(l)("zone-egresses.routes.item.breadcrumbs")}]},{default:m(()=>[z("div",C,[o.value?(e(),t(x,{key:0})):r.value!==null?(e(),t(N,{key:1,error:r.value},null,8,["error"])):s.value===null?(e(),t(A,{key:2})):(e(),E("div",D,[i(b,{"zone-egress-overview":s.value},null,8,["zone-egress-overview"])]))])]),_:1},8,["breadcrumbs"])]),_:1}))}});export{S as default};
