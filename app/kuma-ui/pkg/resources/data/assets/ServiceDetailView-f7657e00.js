import{u as T}from"./vue-router-d8e03a07.js";import{k}from"./kumaApi-41fb4c57.js";import{Q as q}from"./QueryParameter-70743f73.js";import{u as B}from"./store-15db9444.js";import{_ as L}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-76f6a5d9.js";import{E as A}from"./ErrorBlock-1ecc67e5.js";import{_ as I}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-c9a3fc78.js";import{D as V}from"./DataPlaneList-37604d2f.js";import{S as $}from"./ServiceSummary-61f65508.js";import{d as P,h as F,g as z,e as j,f as R,a as d,b as C,F as G,o as c,r as m,s as O}from"./runtime-dom.esm-bundler-32659b48.js";import"./production-58f5acfb.js";import"./kongponents.es-c2485d1e.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./datadogLogEvents-302eea7b.js";import"./ContentWrapper-37eae07f.js";import"./DataOverview-ca93469b.js";import"./StatusBadge-940bb1dd.js";import"./TagList-fc86b2ea.js";import"./YamlView.vue_vue_type_script_setup_true_lang-33fcd065.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-58ff7286.js";import"./toYaml-4e00099e.js";const Q={class:"component-frame"},W=P({__name:"ServiceDetails",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null},dataPlaneOverviews:{type:Array,required:!1,default:null},dppFilterFields:{type:Object,required:!0},selectedDppName:{type:String,required:!1,default:null}},emits:["load-dataplane-overviews"],setup(f,{emit:w}){const a=f;function e(y,t){var o;(((o=a.service.serviceType)==null?void 0:o.startsWith("gateway"))??!1)||delete t.gateway,w("load-dataplane-overviews",y,t)}return(y,t)=>{var n;return c(),F(G,null,[z("div",Q,[j($,{service:a.service,"external-service":f.externalService},null,8,["service","external-service"])]),R(),a.dataPlaneOverviews!==null?(c(),d(V,{key:0,class:"mt-4","data-plane-overviews":a.dataPlaneOverviews,"dpp-filter-fields":a.dppFilterFields,"selected-dpp-name":a.selectedDppName,"is-gateway-view":((n=a.dataPlaneOverviews[0])==null?void 0:n.dataplane.networking.gateway)!==void 0,onLoadData:e},null,8,["data-plane-overviews","dpp-filter-fields","selected-dpp-name","is-gateway-view"])):C("",!0)],64)}}}),J={class:"service-details"},de=P({__name:"ServiceDetailView",props:{selectedDppName:{type:String,required:!1,default:null}},setup(f){const w=f,a={name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},e=T(),y=B(),t=m(null),n=m(null),o=m(null),_=m(!0),g=m(null);O(()=>e.params.mesh,function(){e.name==="service-detail-view"&&h(0)}),O(()=>e.params.name,function(){e.name==="service-detail-view"&&h(0)});function N(){y.dispatch("updatePageTitle",e.params.service);const r=q.get("filterFields"),l=r!==null?JSON.parse(r):{};h(0,l)}N();async function h(r,l={}){_.value=!0,g.value=null,t.value=null,n.value=null,o.value=null;const p=e.params.mesh,v=e.params.service;try{t.value=await k.getServiceInsight({mesh:p,name:v}),t.value.serviceType==="external"?n.value=await k.getExternalServiceByServiceInsightName(p,v):await x(r,l)}catch(s){s instanceof Error?g.value=s:console.error(s)}finally{_.value=!1}}async function x(r,l){const p=e.params.mesh,v=e.params.service;try{const s=b(v,r,l),i=await k.getAllDataplaneOverviewsFromMesh({mesh:p},s);o.value=i.items??[]}catch{o.value=null}}function b(r,l,p){const s=`kuma.io/service:${r}`,i={...p,offset:l,size:50};if(i.tag){const S=Array.isArray(i.tag)?i.tag:[i.tag],D=[];for(const[u,E]of S.entries())E.startsWith("kuma.io/service:")&&D.push(u);for(let u=D.length-1;u===0;u--)S.splice(D[u],1);i.tag=S.concat(s)}else i.tag=s;return i}return(r,l)=>(c(),F("div",J,[_.value?(c(),d(I,{key:0})):g.value!==null?(c(),d(A,{key:1,error:g.value},null,8,["error"])):t.value===null?(c(),d(L,{key:2})):(c(),d(W,{key:3,service:t.value,"data-plane-overviews":o.value,"external-service":n.value,"dpp-filter-fields":a,"selected-dpp-name":w.selectedDppName,onLoadDataplaneOverviews:x},null,8,["service","data-plane-overviews","external-service","selected-dpp-name"]))]))}});export{de as default};
