import{d as _,u as d,r as i,v as u,o as r,j as c,b as l,g}from"./index-04ca1e16.js";import{_ as k}from"./ZoneEgressDetails.vue_vue_type_script_setup_true_lang-952b1dc7.js";import{_ as w}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-9c3782af.js";import{E}from"./ErrorBlock-cd18e62e.js";import{_ as z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-744d045c.js";import{u as h}from"./store-34c42dad.js";import{u as y}from"./index-24620e6e.js";import"./AccordionList-0a04e785.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./DefinitionListItem-53b96571.js";import"./EnvoyData-0a1d188b.js";import"./kongponents.es-bec0b636.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-8260ebe8.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-d6db3d5d.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-9ea279c7.js";import"./TabsWidget-924fc433.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-f525338a.js";const B={class:"zone-details"},$={key:3,class:"kcard-border"},H=_({__name:"ZoneEgressDetailView",setup(b){const p=y(),e=d(),f=h(),o=i(null),t=i(!0),a=i(null);u(()=>e.params.mesh,function(){e.name==="zone-egress-detail-view"&&n()}),u(()=>e.params.name,function(){e.name==="zone-egress-detail-view"&&n()}),v();function v(){f.dispatch("updatePageTitle",e.params.zoneEgress),n()}async function n(){t.value=!0,a.value=null;const m=e.params.zoneEgress;try{o.value=await p.getZoneEgressOverview({name:m})}catch(s){o.value=null,s instanceof Error?a.value=s:console.error(s)}finally{t.value=!1}}return(m,s)=>(r(),c("div",B,[t.value?(r(),l(z,{key:0})):a.value!==null?(r(),l(E,{key:1,error:a.value},null,8,["error"])):o.value===null?(r(),l(w,{key:2})):(r(),c("div",$,[g(k,{"zone-egress-overview":o.value},null,8,["zone-egress-overview"])]))]))}});export{H as default};
