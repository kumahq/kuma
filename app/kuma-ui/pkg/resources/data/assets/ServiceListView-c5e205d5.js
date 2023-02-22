import{u as E}from"./production-fd556ff4.js";import{Q as g}from"./QueryParameter-70743f73.js";import{C}from"./ContentWrapper-645f064f.js";import{D as q}from"./DataOverview-6570508d.js";import{S as L}from"./ServiceSummary-37785e04.js";import{u as R}from"./index-8ca1d18e.js";import{d as V,r,q as z,a as T,w as k,o as P,e as B,b as M}from"./runtime-dom.esm-bundler-062436f2.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./kongponents.es-79677c68.js";import"./datadogLogEvents-302eea7b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-01ce080a.js";import"./ErrorBlock-f54abf42.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-d0e41aae.js";import"./StatusBadge-5a5c160e.js";import"./TagList-59ee9bba.js";import"./YamlView.vue_vue_type_script_setup_true_lang-b4e1e79e.js";import"./store-59c1cb7c.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-eaa82ee4.js";import"./toYaml-4e00099e.js";const se=V({__name:"ServiceListView",props:{selectedServiceName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(_){const m=_,c=R(),A=[{label:"Service",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],S=50,N={title:"No Data",message:"There are no service insights present."},u=E(),p=r(!0),v=r(null),x=r(null),b=r(m.offset),o=r(null),h=r(null),l=r({headers:A,data:[]});z(()=>u.params.mesh,function(){u.name==="service-list-view"&&d(0)}),d(m.offset);async function d(e){b.value=e,g.set("offset",e>0?e:null),p.value=!0,v.value=null;const t=u.params.mesh,s=S;try{const{items:a,next:f}=await c.getAllServiceInsightsFromMesh({mesh:t},{size:s,offset:e});if(x.value=f,Array.isArray(a)&&a.length>0){a.sort((n,i)=>n.name>i.name?1:n.name<i.name?-1:0),l.value.data=a.map(n=>D(n));const y=m.selectedServiceName??a[0].name;await w({name:y,mesh:t})}else l.value.data=[]}catch(a){a instanceof Error?v.value=a:console.error(a)}finally{p.value=!1}}function D(e){const t={name:"service-detail-view",params:{mesh:e.mesh,service:e.name}},s={name:"mesh-detail-view",params:{mesh:e.mesh}};let a="—";if(e.dataplanes){const{online:n=0,total:i=0}=e.dataplanes;a=`${n} / ${i}`}const f=e.addressPort,y=e.serviceType??"internal";return{...e,serviceType:y,nameRoute:t,meshRoute:s,dpProxiesStatus:a,addressPort:f}}async function w({mesh:e,name:t}){o.value=await c.getServiceInsight({mesh:e,name:t}),o.value.serviceType==="external"&&(h.value=await c.getExternalServiceByServiceInsightName(e,t)),g.set("service",t)}return(e,t)=>(P(),T(C,null,{content:k(()=>{var s;return[B(q,{"selected-entity-name":(s=o.value)==null?void 0:s.name,"page-size":S,error:v.value,"is-loading":p.value,"empty-state":N,"table-data":l.value,"table-data-is-empty":l.value.data.length===0,next:x.value,"page-offset":b.value,onTableAction:w,onLoadData:d},null,8,["selected-entity-name","error","is-loading","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:k(()=>[o.value!==null?(P(),T(L,{key:0,service:o.value,"external-service":h.value},null,8,["service","external-service"])):M("",!0)]),_:1}))}});export{se as default};
