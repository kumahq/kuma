import{d as _,u as d,r as i,v as u,o as r,j as c,b as l,g}from"./index-987a13b5.js";import{_ as k}from"./ZoneEgressDetails.vue_vue_type_script_setup_true_lang-416c74ac.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-846e8989.js";import{E}from"./ErrorBlock-e418ef19.js";import{_ as z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-9d99661c.js";import{u as h}from"./store-b4114847.js";import{u as y}from"./index-6dab244f.js";import"./AccordionList-51e44c1b.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DefinitionListItem-1cd1355b.js";import"./EnvoyData-2c3b288c.js";import"./kongponents.es-7e482eb5.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2144e5ca.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-c94f127a.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-a3b2270c.js";import"./TabsWidget-89c9f343.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-d256feea.js";const B={class:"zone-details"},$={key:3,class:"kcard-border"},H=_({__name:"ZoneEgressDetailView",setup(b){const p=y(),e=d(),f=h(),o=i(null),t=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-egress-detail-view"&&n()}),u(()=>e.params.name,function(){e.name==="zone-egress-detail-view"&&n()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneEgress),n()}async function n(){t.value=!0,a.value=null;const m=e.params.zoneEgress;try{o.value=await p.getZoneEgressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{t.value=!1}}return(m,s)=>(r(),c("div",B,[t.value?(r(),l(z,{key:0})):a.value!==null?(r(),l(E,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",$,[g(k,{"zone-egress-overview":o.value},null,8,["zone-egress-overview"])]))]))}});export{H as default};
