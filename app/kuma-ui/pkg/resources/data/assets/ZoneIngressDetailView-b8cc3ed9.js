import{d as _,u as d,r as i,v as u,j as c,e as l,g,o as r}from"./index-a24b4f04.js";import{_ as k}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-76d16d90.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-8ed542ca.js";import{E as z}from"./ErrorBlock-8e1d70a5.js";import{_ as h}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-4f020979.js";import{u as y}from"./store-07fabdaf.js";import{u as I}from"./index-f7ac63b4.js";import"./AccordionList-8e75dbac.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./EnvoyData-7aa4844b.js";import"./kongponents.es-5adaddec.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-7b080a3a.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-b3124157.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-f02238bb.js";import"./TabsWidget-6a5a1765.js";import"./datadogLogEvents-302eea7b.js";import"./QueryParameter-70743f73.js";const B={class:"zone-details"},E={key:3,class:"kcard-border"},F=_({__name:"ZoneIngressDetailView",setup($){const p=I(),e=d(),f=y(),o=i(null),n=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-ingress-detail-view"&&t()}),u(()=>e.params.name,function(){e.name==="zone-ingress-detail-view"&&t()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneIngress),t()}async function t(){n.value=!0,a.value=null;const m=e.params.zoneIngress;try{o.value=await p.getZoneIngressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{n.value=!1}}return(m,s)=>(r(),c("div",B,[n.value?(r(),l(h,{key:0})):a.value!==null?(r(),l(z,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",E,[g(k,{"zone-ingress-overview":o.value},null,8,["zone-ingress-overview"])]))]))}});export{F as default};
