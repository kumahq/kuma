import{d as T,u as A,r as o,v as D,y as x,A as Z,j as _,i as E,g as h,b as V,h as B,f as S,B as b,o as w}from"./index-bd38c154.js";import{_ as I}from"./ZoneEgressDetails.vue_vue_type_script_setup_true_lang-27614e15.js";import{D as L}from"./DataOverview-d216744b.js";import{u as O}from"./index-2e645818.js";import{Q as d}from"./QueryParameter-70743f73.js";import"./AccordionList-68fd7c69.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-d6e052e1.js";import"./kongponents.es-ba82ceca.js";import"./EnvoyData-3e492041.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-39ef11a2.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-12b7659b.js";import"./ErrorBlock-99d9a9e3.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-fc8e76df.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-22eb5fc3.js";import"./TabsWidget-59c9beec.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-fcd7936c.js";import"./store-82b3ee45.js";import"./StatusBadge-fac45f76.js";const P={class:"zoneegresses"},q={class:"kcard-stack"},C={class:"kcard-border"},F={key:0,class:"kcard-border"},le=T({__name:"ZoneEgressListView",props:{selectedZoneEgressName:{type:[String,null],required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(z){const l=z,p=O(),k={title:"No Data",message:"There are no Zone Egresses present."},v=A(),i=o(!0),u=o(null),n=o({headers:[{label:"Status",key:"status"},{label:"Name",key:"entity"}],data:[]}),r=o(null),f=o(null),g=o(l.offset);D(()=>v.params.mesh,function(){v.name==="zone-egress-list-view"&&c(0)}),x(function(){c(l.offset)});async function c(e){var a;g.value=e,d.set("offset",e>0?e:null),i.value=!0,u.value=null;const t=b;try{const{items:s,next:m}=await p.getAllZoneEgressOverviews({size:t,offset:e});f.value=m,n.value.data=N(s??[]),await y({name:l.selectedZoneEgressName??((a=n.value.data[0])==null?void 0:a.entity.name)})}catch(s){n.value.data=[],r.value=null,s instanceof Error?u.value=s:console.error(s)}finally{i.value=!1}}function N(e){return e.map(t=>{const{name:a}=t,s={name:"zone-egress-detail-view",params:{zoneEgress:a}},m=Z(t.zoneEgressInsight??{});return{entity:t,detailViewRoute:s,status:m}})}async function y({name:e}){if(e===void 0){r.value=null,d.set("zoneEgress",null);return}try{r.value=await p.getZoneEgressOverview({name:e}),d.set("zoneEgress",e)}catch(t){console.error(t)}}return(e,t)=>{var a;return w(),_("div",P,[E("div",q,[E("div",C,[h(L,{"selected-entity-name":(a=r.value)==null?void 0:a.name,"page-size":V(b),"is-loading":i.value,error:u.value,"empty-state":k,"table-data":n.value,"table-data-is-empty":n.value.data.length===0,next:f.value,"page-offset":g.value,onTableAction:y,onLoadData:c},null,8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]),B(),r.value!==null?(w(),_("div",F,[h(I,{"zone-egress-overview":r.value},null,8,["zone-egress-overview"])])):S("",!0)])])}}});export{le as default};
