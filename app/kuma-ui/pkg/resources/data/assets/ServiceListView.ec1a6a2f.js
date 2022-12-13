import{d as P,p as C,r as l,g as E,k as N,cA as q,o as T,c as x,w as A,a as L,A as R}from"./index.60b0f0ac.js";import{C as V}from"./ContentWrapper.1618bbe9.js";import{p as z,D as M}from"./patchQueryParam.ae688d93.js";import{S as O}from"./ServiceDetails.38c00505.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.548da37c.js";import"./EntityStatus.6fc3c7d6.js";import"./ErrorBlock.2ee4d08e.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.be7e4bb1.js";import"./TagList.b3d2d71f.js";import"./YamlView.vue_vue_type_script_setup_true_lang.152633f3.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.85e36160.js";import"./_commonjsHelpers.f037b798.js";const X=P({__name:"ServiceListView",props:{offset:{type:Number,required:!1,default:0}},setup(g){const f=g,k=[{label:"Service",key:"name"},{label:"Type",key:"serviceType"},{label:"Address",key:"address"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],y=50,w={title:"No Data",message:"There are no service insights present."},o=C(),p=l(!0),v=l(null),h=l(null),S=l(f.offset),s=l(null),i=l({headers:k,data:[]});E(()=>o.params.mesh,function(){o.name==="service-list-view"&&d(0)}),d(f.offset);async function d(e){S.value=e,z("offset",e>0?e:null),p.value=!0,v.value=null;const u=o.params.mesh,r=y;try{const{items:a,next:m}=await N.getAllServiceInsightsFromMesh({mesh:u},{size:r,offset:e});if(h.value=m,Array.isArray(a)&&a.length>0){a.sort((t,n)=>t.name>n.name?1:t.name<n.name?-1:0);let c=a[0];if(o.query.ns){const t=a.find(n=>n.name===o.query.ns);t!==void 0&&(c=t)}b(c),i.value.data=a.map(t=>_(t))}else s.value=null,i.value.data=[]}catch(a){s.value=null,a instanceof Error?v.value=a:console.error(a)}finally{p.value=!1}}function _(e){var t;const u={name:e.serviceType==="external"?"external-service-detail-view":"service-insight-detail-view",params:{mesh:e.mesh,service:e.name}},r={name:"mesh-detail-view",params:{mesh:e.mesh}};let a="\u2014";if(e.dataplanes){const{online:n=0,total:D=0}=e.dataplanes;a=`${n} / ${D}`}let m="\u2014";e.status&&(m=q[e.status].title);const c=(t=e.serviceType)!=null?t:"internal";return{...e,serviceType:c,nameRoute:u,meshRoute:r,dpProxiesStatus:a,status:m}}function b(e){s.value=e}return(e,u)=>(T(),x(V,null,{content:A(()=>{var r;return[L(M,{"selected-entity-name":(r=s.value)==null?void 0:r.name,"page-size":y,error:v.value,"is-loading":p.value,"empty-state":w,"table-data":i.value,"table-data-is-empty":i.value.data.length===0,next:h.value,"page-offset":S.value,onTableAction:b,onLoadData:d},null,8,["selected-entity-name","error","is-loading","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:A(()=>[s.value!==null?(T(),x(O,{key:0,name:s.value.name,mesh:s.value.mesh,"service-type":s.value.serviceType},null,8,["name","mesh","service-type"])):R("",!0)]),_:1}))}});export{X as default};
