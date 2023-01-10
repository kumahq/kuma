import{d as C,q,r,f as L,k as S,o as k,c as T,w as P,a as R,A as V}from"./index.0c4c6d47.js";import{C as z}from"./ContentWrapper.0535eefe.js";import{p as _,D as M}from"./patchQueryParam.3ef0b93e.js";import{S as O}from"./ServiceSummary.f2874919.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.16c6ba85.js";import"./ErrorBlock.39782751.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.e84843f4.js";import"./StatusBadge.e9f457ff.js";import"./TagList.0486290a.js";import"./YamlView.vue_vue_type_script_setup_true_lang.812a5c39.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.cc0862d4.js";import"./_commonjsHelpers.f037b798.js";const X=C({__name:"ServiceListView",props:{selectedServiceName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(A){const m=A,D=[{label:"Service",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"addressPort"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],x=50,N={title:"No Data",message:"There are no service insights present."},p=q(),v=r(!0),d=r(null),h=r(null),b=r(m.offset),l=r(null),w=r(null),i=r({headers:D,data:[]});L(()=>p.params.mesh,function(){p.name==="service-list-view"&&f(0)}),f(m.offset);async function f(e){var o;b.value=e,_("offset",e>0?e:null),v.value=!0,d.value=null;const t=p.params.mesh,s=x;try{const{items:a,next:y}=await S.getAllServiceInsightsFromMesh({mesh:t},{size:s,offset:e});if(h.value=y,Array.isArray(a)&&a.length>0){a.sort((n,u)=>n.name>u.name?1:n.name<u.name?-1:0),i.value.data=a.map(n=>E(n));const c=(o=m.selectedServiceName)!=null?o:a[0].name;await g({name:c,mesh:t})}else i.value.data=[]}catch(a){a instanceof Error?d.value=a:console.error(a)}finally{v.value=!1}}function E(e){var c;const t={name:"service-detail-view",params:{mesh:e.mesh,service:e.name}},s={name:"mesh-detail-view",params:{mesh:e.mesh}};let o="\u2014";if(e.dataplanes){const{online:n=0,total:u=0}=e.dataplanes;o=`${n} / ${u}`}const a=e.addressPort,y=(c=e.serviceType)!=null?c:"internal";return{...e,serviceType:y,nameRoute:t,meshRoute:s,dpProxiesStatus:o,addressPort:a}}async function g({mesh:e,name:t}){l.value=await S.getServiceInsight({mesh:e,name:t}),l.value.serviceType==="external"&&(w.value=await S.getExternalService({mesh:e,name:t})),_("service",t)}return(e,t)=>(k(),T(z,null,{content:P(()=>{var s;return[R(M,{"selected-entity-name":(s=l.value)==null?void 0:s.name,"page-size":x,error:d.value,"is-loading":v.value,"empty-state":N,"table-data":i.value,"table-data-is-empty":i.value.data.length===0,next:h.value,"page-offset":b.value,onTableAction:g,onLoadData:f},null,8,["selected-entity-name","error","is-loading","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:P(()=>[l.value!==null?(k(),T(O,{key:0,service:l.value,"external-service":w.value},null,8,["service","external-service"])):V("",!0)]),_:1}))}});export{X as default};
