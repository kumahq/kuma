import{d as _,u as d,r as i,v as c,o,j as u,b as l,g as k}from"./index-22d1354e.js";import{_ as w}from"./ZoneDetails.vue_vue_type_script_setup_true_lang-ef8abe35.js";import{_ as z}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-7090adea.js";import{E as h}from"./ErrorBlock-cc03b6cb.js";import{_ as y}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-9054bd94.js";import{u as g}from"./store-10af0af9.js";import{u as B}from"./index-18f05561.js";import"./kongponents.es-14a71563.js";import"./AccordionList-6865d97f.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ed00a067.js";import"./DefinitionListItem-fe984ffc.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-dae6cd59.js";import"./TabsWidget-317324b0.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-df204c95.js";import"./WarningsWidget.vue_vue_type_script_setup_true_lang-b3715d2a.js";const E={class:"zone-details"},$={key:3,class:"kcard-border"},G=_({__name:"ZoneDetailView",setup(b){const p=B(),e=d(),f=g(),a=i(null),n=i(!0),r=i(null);c(()=>e.params.mesh,function(){e.name==="zone-cp-detail-view"&&s()}),c(()=>e.params.name,function(){e.name==="zone-cp-detail-view"&&s()}),v();function v(){f.dispatch("updatePageTitle",e.params.zone),s()}async function s(){n.value=!0,r.value=null;const m=e.params.zone;try{a.value=await p.getZoneOverview({name:m})}catch(t){a.value=null,t instanceof Error?r.value=t:console.error(t)}finally{n.value=!1}}return(m,t)=>(o(),u("div",E,[n.value?(o(),l(y,{key:0})):r.value!==null?(o(),l(h,{key:1,error:r.value},null,8,["error"])):a.value===null?(o(),l(z,{key:2})):(o(),u("div",$,[k(w,{"zone-overview":a.value},null,8,["zone-overview"])]))]))}});export{G as default};
