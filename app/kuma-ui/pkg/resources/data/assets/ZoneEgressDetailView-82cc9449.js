import{d as _,u as d,r as i,v as u,j as c,b as l,g,o as r}from"./index-0fbacd76.js";import{_ as k}from"./ZoneEgressDetails.vue_vue_type_script_setup_true_lang-ecb84547.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-c0c5c459.js";import{E}from"./ErrorBlock-09ceca3c.js";import{_ as z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-3d8f32b0.js";import{u as h}from"./store-5ee7e2bf.js";import{u as y}from"./index-bc240554.js";import"./AccordionList-1e0cd60d.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DefinitionListItem-87124c14.js";import"./EnvoyData-3b512034.js";import"./kongponents.es-c741eab8.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-90b7b8c9.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-de86e759.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-dceeb786.js";import"./TabsWidget-597e6684.js";import"./datadogLogEvents-302eea7b.js";import"./QueryParameter-70743f73.js";const B={class:"zone-details"},$={key:3,class:"kcard-border"},H=_({__name:"ZoneEgressDetailView",setup(b){const p=y(),e=d(),f=h(),o=i(null),t=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-egress-detail-view"&&n()}),u(()=>e.params.name,function(){e.name==="zone-egress-detail-view"&&n()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneEgress),n()}async function n(){t.value=!0,a.value=null;const m=e.params.zoneEgress;try{o.value=await p.getZoneEgressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{t.value=!1}}return(m,s)=>(r(),c("div",B,[t.value?(r(),l(z,{key:0})):a.value!==null?(r(),l(E,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",$,[g(k,{"zone-egress-overview":o.value},null,8,["zone-egress-overview"])]))]))}});export{H as default};
