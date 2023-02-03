import{u as T}from"./vue-router-573afc44.js";import{k,u as q}from"./store-713e15c3.js";import{Q as B}from"./QueryParameter-70743f73.js";import{_ as L}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-4047971f.js";import{E as A}from"./ErrorBlock-940f5e66.js";import{_ as I}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-d3176fee.js";import{D as V}from"./DataPlaneList-76555925.js";import{S as $}from"./ServiceSummary-a2856ca0.js";import{d as P,h as F,g as z,e as j,f as R,a as d,b as C,F as G,o as c,r as m,y as O}from"./runtime-dom.esm-bundler-91b41870.js";import"./vuex.esm-bundler-df5bd11e.js";import"./constants-31fdaf55.js";import"./kongponents.es-3df60cd6.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./datadogLogEvents-4578cfa7.js";import"./ContentWrapper-b0fa1a61.js";import"./DataOverview-ee1e202c.js";import"./StatusBadge-81464ebd.js";import"./TagList-91d1133a.js";import"./YamlView.vue_vue_type_script_setup_true_lang-793819c9.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-d3e1ee08.js";import"./_commonjsHelpers-87174ba5.js";import"./toYaml-4e00099e.js";const Q={class:"component-frame"},W=P({__name:"ServiceDetails",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null},dataPlaneOverviews:{type:Array,required:!1,default:null},dppFilterFields:{type:Object,required:!0},selectedDppName:{type:String,required:!1,default:null}},emits:["load-dataplane-overviews"],setup(f,{emit:w}){const a=f;function e(y,t){var o;(((o=a.service.serviceType)==null?void 0:o.startsWith("gateway"))??!1)||delete t.gateway,w("load-dataplane-overviews",y,t)}return(y,t)=>{var n;return c(),F(G,null,[z("div",Q,[j($,{service:a.service,"external-service":f.externalService},null,8,["service","external-service"])]),R(),a.dataPlaneOverviews!==null?(c(),d(V,{key:0,class:"mt-4","data-plane-overviews":a.dataPlaneOverviews,"dpp-filter-fields":a.dppFilterFields,"selected-dpp-name":a.selectedDppName,"is-gateway-view":((n=a.dataPlaneOverviews[0])==null?void 0:n.dataplane.networking.gateway)!==void 0,onLoadData:e},null,8,["data-plane-overviews","dpp-filter-fields","selected-dpp-name","is-gateway-view"])):C("",!0)],64)}}}),J={class:"service-details"},fe=P({__name:"ServiceDetailView",props:{selectedDppName:{type:String,required:!1,default:null}},setup(f){const w=f,a={name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},e=T(),y=q(),t=m(null),n=m(null),o=m(null),_=m(!0),g=m(null);O(()=>e.params.mesh,function(){e.name==="service-detail-view"&&h(0)}),O(()=>e.params.name,function(){e.name==="service-detail-view"&&h(0)});function N(){y.dispatch("updatePageTitle",e.params.service);const r=B.get("filterFields"),l=r!==null?JSON.parse(r):{};h(0,l)}N();async function h(r,l={}){_.value=!0,g.value=null,t.value=null,n.value=null,o.value=null;const p=e.params.mesh,v=e.params.service;try{t.value=await k.getServiceInsight({mesh:p,name:v}),t.value.serviceType==="external"?n.value=await k.getExternalServiceByServiceInsightName(p,v):await x(r,l)}catch(s){s instanceof Error?g.value=s:console.error(s)}finally{_.value=!1}}async function x(r,l){const p=e.params.mesh,v=e.params.service;try{const s=b(v,r,l),i=await k.getAllDataplaneOverviewsFromMesh({mesh:p},s);o.value=i.items??[]}catch{o.value=null}}function b(r,l,p){const s=`kuma.io/service:${r}`,i={...p,offset:l,size:50};if(i.tag){const S=Array.isArray(i.tag)?i.tag:[i.tag],D=[];for(const[u,E]of S.entries())E.startsWith("kuma.io/service:")&&D.push(u);for(let u=D.length-1;u===0;u--)S.splice(D[u],1);i.tag=S.concat(s)}else i.tag=s;return i}return(r,l)=>(c(),F("div",J,[_.value?(c(),d(I,{key:0})):g.value!==null?(c(),d(A,{key:1,error:g.value},null,8,["error"])):t.value===null?(c(),d(L,{key:2})):(c(),d(W,{key:3,service:t.value,"data-plane-overviews":o.value,"external-service":n.value,"dpp-filter-fields":a,"selected-dpp-name":w.selectedDppName,onLoadDataplaneOverviews:x},null,8,["service","data-plane-overviews","external-service","selected-dpp-name"]))]))}});export{fe as default};
