import{d as _,u as d,r as i,v as u,o as r,j as c,b as l,g}from"./index-6f18d0d5.js";import{_ as k}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-1e7def66.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-57391115.js";import{E as z}from"./ErrorBlock-b14b6241.js";import{_ as h}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-646c9f8e.js";import{u as y}from"./store-ffbd1d65.js";import{u as I}from"./index-ba9458d7.js";import"./AccordionList-3b91dfed.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DefinitionListItem-8bc09ece.js";import"./EnvoyData-974e9e04.js";import"./kongponents.es-b7114d58.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-d9734676.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-e7fa6724.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-339cb323.js";import"./TabsWidget-e5fe1ef9.js";import"./datadogLogEvents-302eea7b.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-fe762ac6.js";const B={class:"zone-details"},E={key:3,class:"kcard-border"},H=_({__name:"ZoneIngressDetailView",setup($){const p=I(),e=d(),f=y(),o=i(null),t=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-ingress-detail-view"&&n()}),u(()=>e.params.name,function(){e.name==="zone-ingress-detail-view"&&n()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneIngress),n()}async function n(){t.value=!0,a.value=null;const m=e.params.zoneIngress;try{o.value=await p.getZoneIngressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{t.value=!1}}return(m,s)=>(r(),c("div",B,[t.value?(r(),l(h,{key:0})):a.value!==null?(r(),l(z,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",E,[g(k,{"zone-ingress-overview":o.value},null,8,["zone-ingress-overview"])]))]))}});export{H as default};
