import{d as _,u as d,r as i,v as u,o as r,j as c,b as l,g}from"./index-61cef882.js";import{_ as k}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-38afb3b8.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-5a7795a6.js";import{E as z}from"./ErrorBlock-e115e1aa.js";import{_ as h}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-fa2a2bb6.js";import{u as y}from"./store-0af1ff9f.js";import{u as I}from"./index-01e79acb.js";import"./AccordionList-039f6d8c.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DefinitionListItem-841dbb71.js";import"./EnvoyData-796de2c1.js";import"./kongponents.es-d381709c.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-4f359e86.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-6ccffd0b.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-f40ad698.js";import"./TabsWidget-e9cec750.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-037745d5.js";const B={class:"zone-details"},E={key:3,class:"kcard-border"},G=_({__name:"ZoneIngressDetailView",setup($){const p=I(),e=d(),f=y(),o=i(null),n=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-ingress-detail-view"&&t()}),u(()=>e.params.name,function(){e.name==="zone-ingress-detail-view"&&t()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneIngress),t()}async function t(){n.value=!0,a.value=null;const m=e.params.zoneIngress;try{o.value=await p.getZoneIngressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{n.value=!1}}return(m,s)=>(r(),c("div",B,[n.value?(r(),l(h,{key:0})):a.value!==null?(r(),l(z,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",E,[g(k,{"zone-ingress-overview":o.value},null,8,["zone-ingress-overview"])]))]))}});export{G as default};
