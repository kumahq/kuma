import{d as q,o,f as B,a as k,b as E,c as p,g as I,F as L,q as V,r as d,s as P,w as F,u as N,e as z}from"./index-271b6183.js";import{S as j}from"./ServiceSummary-79407c48.js";import{D as C}from"./DataPlaneList-e1bd9274.js";import{u as R,b as G,g as Q,f as W,e as J}from"./RouteView.vue_vue_type_script_setup_true_lang-fadf0571.js";import{_ as K}from"./RouteTitle.vue_vue_type_script_setup_true_lang-1f8fd421.js";import{_ as M}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-9d9f8054.js";import{E as H}from"./ErrorBlock-2e363ab2.js";import{_ as U}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-c91f8087.js";import{Q as X}from"./QueryParameter-70743f73.js";import"./kongponents.es-dc880404.js";import"./DefinitionListItem-cf681f64.js";import"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-cfdfffdc.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ac8fa1e9.js";import"./TextWithCopyButton-3ab4305e.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-85609c66.js";import"./toYaml-4e00099e.js";import"./StatusBadge-2721d158.js";import"./TagList-5e49a7a3.js";import"./ContentWrapper-bedb19e3.js";import"./DataOverview-d4799d67.js";const Y=q({__name:"ServiceDetails",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null},dataPlaneOverviews:{type:Array,required:!1,default:null},dppFilterFields:{type:Object,required:!0},selectedDppName:{type:String,required:!1,default:null}},emits:["load-dataplane-overviews"],setup(f,{emit:_}){const e=f;function a(w,c){var l;(((l=e.service.serviceType)==null?void 0:l.startsWith("gateway"))??!1)||delete c.gateway,_("load-dataplane-overviews",w,c)}return(w,c)=>{var u;return o(),B(L,null,[k(j,{service:e.service,"external-service":f.externalService},null,8,["service","external-service"]),E(),e.dataPlaneOverviews!==null?(o(),p(C,{key:0,class:"mt-4","data-plane-overviews":e.dataPlaneOverviews,"dpp-filter-fields":e.dppFilterFields,"selected-dpp-name":e.selectedDppName,"is-gateway-view":((u=e.dataPlaneOverviews[0])==null?void 0:u.dataplane.networking.gateway)!==void 0,onLoadData:a},null,8,["data-plane-overviews","dpp-filter-fields","selected-dpp-name","is-gateway-view"])):I("",!0)],64)}}}),Z={class:"service-details"},he=q({__name:"ServiceDetailView",props:{selectedDppName:{type:String,required:!1,default:null}},setup(f){const _=f,e=R(),a=V(),w=G(),{t:c}=Q(),u={name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},l=d(null),h=d(null),g=d(null),S=d(!0),y=d(null);P(()=>a.params.mesh,function(){a.name==="service-detail-view"&&x(0)}),P(()=>a.params.name,function(){a.name==="service-detail-view"&&x(0)});function T(){w.dispatch("updatePageTitle",a.params.service);const t=X.get("filterFields"),n=t!==null?JSON.parse(t):{};x(0,n)}T();async function x(t,n={}){S.value=!0,y.value=null,l.value=null,h.value=null,g.value=null;const r=a.params.mesh,v=a.params.service;try{l.value=await e.getServiceInsight({mesh:r,name:v}),l.value.serviceType==="external"?h.value=await e.getExternalServiceByServiceInsightName(r,v):await O(t,n)}catch(s){s instanceof Error?y.value=s:console.error(s)}finally{S.value=!1}}async function O(t,n){const r=a.params.mesh,v=a.params.service;try{const s=$(v,t,n),i=await e.getAllDataplaneOverviewsFromMesh({mesh:r},s);g.value=i.items??[]}catch{g.value=null}}function $(t,n,r){const s=`kuma.io/service:${t}`,i={...r,offset:n,size:50};if(i.tag){const D=Array.isArray(i.tag)?i.tag:[i.tag],b=[];for(const[m,A]of D.entries())A.startsWith("kuma.io/service:")&&b.push(m);for(let m=b.length-1;m===0;m--)D.splice(b[m],1);i.tag=D.concat(s)}else i.tag=s;return i}return(t,n)=>(o(),p(J,null,{default:F(({route:r})=>[k(K,{title:N(c)("services.routes.item.title",{name:r.params.service})},null,8,["title"]),E(),k(W,{breadcrumbs:[{to:{name:"services-list-view",params:r.params},text:N(c)("services.routes.item.breadcrumbs")}]},{default:F(()=>[z("div",Z,[S.value?(o(),p(U,{key:0})):y.value!==null?(o(),p(H,{key:1,error:y.value},null,8,["error"])):l.value===null?(o(),p(M,{key:2})):(o(),p(Y,{key:3,service:l.value,"data-plane-overviews":g.value,"external-service":h.value,"dpp-filter-fields":u,"selected-dpp-name":_.selectedDppName,onLoadDataplaneOverviews:O},null,8,["service","data-plane-overviews","external-service","selected-dpp-name"]))])]),_:2},1032,["breadcrumbs"])]),_:1}))}});export{he as default};
