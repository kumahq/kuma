import{d as k,u as w,q as n,o as e,a as r,w as c,h as i,b as m,g as z,k as b,e as h}from"./index-abe682b3.js";import{_ as y}from"./ZoneDetails.vue_vue_type_script_setup_true_lang-26c9ca5a.js";import{g as x,i as B,h as V,A as $,_ as g}from"./RouteView.vue_vue_type_script_setup_true_lang-99401e5a.js";import{_ as A}from"./RouteTitle.vue_vue_type_script_setup_true_lang-64e3d4a6.js";import{_ as E}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-147f18f2.js";import{E as N}from"./ErrorBlock-fe291eb6.js";import"./kongponents.es-406a7d3e.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-9565d7a4.js";import"./DefinitionListItem-aa0345b2.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-c332ba11.js";import"./TabsWidget-39cc07d3.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-a3aee751.js";import"./WarningsWidget.vue_vue_type_script_setup_true_lang-715ba15e.js";const C={class:"zone-details"},D={key:3,class:"kcard-border","data-testid":"detail-view-details"},Q=k({__name:"ZoneDetailView",setup(O){const p=x(),_=w(),{t:l}=B(),a=n(null),s=n(!0),t=n(null);d();function d(){f()}async function f(){s.value=!0,t.value=null;const u=_.params.zone;try{a.value=await p.getZoneOverview({name:u})}catch(o){a.value=null,o instanceof Error?t.value=o:console.error(o)}finally{s.value=!1}}return(u,o)=>(e(),r(g,null,{default:c(({route:v})=>[i(A,{title:m(l)("zone-cps.routes.item.title",{name:v.params.zone})},null,8,["title"]),z(),i($,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:m(l)("zone-cps.routes.item.breadcrumbs")}]},{default:c(()=>[b("div",C,[s.value?(e(),r(V,{key:0})):t.value!==null?(e(),r(N,{key:1,error:t.value},null,8,["error"])):a.value===null?(e(),r(E,{key:2})):(e(),h("div",D,[i(y,{"zone-overview":a.value},null,8,["zone-overview"])]))])]),_:1},8,["breadcrumbs"])]),_:1}))}});export{Q as default};
