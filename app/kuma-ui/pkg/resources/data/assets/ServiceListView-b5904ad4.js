import{u as E}from"./vue-router-573afc44.js";import{k as y}from"./store-f89d97b2.js";import{Q as g}from"./QueryParameter-70743f73.js";import{C}from"./ContentWrapper-b0fa1a61.js";import{D as L}from"./DataOverview-b5021cb1.js";import{S as R}from"./ServiceSummary-828d0097.js";import{d as V,r,y as q,a as k,w as T,o as P,e as z,b as M}from"./runtime-dom.esm-bundler-91b41870.js";import"./vuex.esm-bundler-df5bd11e.js";import"./constants-31fdaf55.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./kongponents.es-3df60cd6.js";import"./datadogLogEvents-4578cfa7.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-4047971f.js";import"./ErrorBlock-4d9ed375.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-d3176fee.js";import"./StatusBadge-81464ebd.js";import"./TagList-91d1133a.js";import"./YamlView.vue_vue_type_script_setup_true_lang-4906a29a.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-28681e3c.js";import"./_commonjsHelpers-87174ba5.js";import"./toYaml-4e00099e.js";const oe=V({__name:"ServiceListView",props:{selectedServiceName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(_){const m=_,A=[{label:"Service",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],S=50,D={title:"No Data",message:"There are no service insights present."},c=E(),u=r(!0),p=r(null),x=r(null),b=r(m.offset),n=r(null),h=r(null),l=r({headers:A,data:[]});q(()=>c.params.mesh,function(){c.name==="service-list-view"&&v(0)}),v(m.offset);async function v(e){b.value=e,g.set("offset",e>0?e:null),u.value=!0,p.value=null;const t=c.params.mesh,s=S;try{const{items:a,next:d}=await y.getAllServiceInsightsFromMesh({mesh:t},{size:s,offset:e});if(x.value=d,Array.isArray(a)&&a.length>0){a.sort((o,i)=>o.name>i.name?1:o.name<i.name?-1:0),l.value.data=a.map(o=>N(o));const f=m.selectedServiceName??a[0].name;await w({name:f,mesh:t})}else l.value.data=[]}catch(a){a instanceof Error?p.value=a:console.error(a)}finally{u.value=!1}}function N(e){const t={name:"service-detail-view",params:{mesh:e.mesh,service:e.name}},s={name:"mesh-detail-view",params:{mesh:e.mesh}};let a="—";if(e.dataplanes){const{online:o=0,total:i=0}=e.dataplanes;a=`${o} / ${i}`}const d=e.addressPort,f=e.serviceType??"internal";return{...e,serviceType:f,nameRoute:t,meshRoute:s,dpProxiesStatus:a,addressPort:d}}async function w({mesh:e,name:t}){n.value=await y.getServiceInsight({mesh:e,name:t}),n.value.serviceType==="external"&&(h.value=await y.getExternalService({mesh:e,name:t})),g.set("service",t)}return(e,t)=>(P(),k(C,null,{content:T(()=>{var s;return[z(L,{"selected-entity-name":(s=n.value)==null?void 0:s.name,"page-size":S,error:p.value,"is-loading":u.value,"empty-state":D,"table-data":l.value,"table-data-is-empty":l.value.data.length===0,next:x.value,"page-offset":b.value,onTableAction:w,onLoadData:v},null,8,["selected-entity-name","error","is-loading","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:T(()=>[n.value!==null?(P(),k(R,{key:0,service:n.value,"external-service":h.value},null,8,["service","external-service"])):M("",!0)]),_:1}))}});export{oe as default};
