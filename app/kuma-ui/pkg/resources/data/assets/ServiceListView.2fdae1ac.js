import{d as D,u as A,j as n,l as C,K as E,cz as P,c as f,w as b,o as S,a as g,p as z}from"./index.f4381a04.js";import{C as L}from"./ContentWrapper.34b274d5.js";import{D as M}from"./DataOverview.2dc14e2d.js";import{S as N}from"./ServiceDetails.0ac93bf3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.a081fa47.js";import"./ErrorBlock.3c391f50.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.52c551fa.js";import"./EntityTag.819c2a1f.js";import"./EntityURLControl.a9be977c.js";import"./YamlView.8ed40dd6.js";import"./index.58caa11d.js";import"./CodeBlock.2eccb873.js";import"./_commonjsHelpers.f037b798.js";const H=D({__name:"ServiceListView",setup(R){const T=[{label:"Service",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"serviceType"},{label:"Address",key:"address"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],d=50,x={title:"No Data",message:"There are no service insights present."},m=A(),c=n(!0),l=n(null),h=n(null),t=n(null),o=n({headers:T,data:[]});C(()=>m.params.mesh,function(){m.name==="service-list-view"&&p(0)}),p(0);async function p(e){c.value=!0,l.value=null;const i=m.params.mesh,v=d;try{const{items:a=[],next:u}=await E.getAllServiceInsightsFromMesh({mesh:i},{size:v,offset:e});h.value=u,Array.isArray(a)&&a.length>0?(a.sort((s,r)=>s.name>r.name?1:s.name<r.name?-1:s.mesh.localeCompare(r.mesh)),y(a[0]),o.value.data=a.map(s=>_(s))):(t.value=null,o.value.data=[])}catch(a){t.value=null,a instanceof Error&&(l.value=a),console.error(l)}finally{c.value=!1}}function _(e){var r;const i={name:e.serviceType==="external"?"external-service-detail-view":"service-insight-detail-view",params:{mesh:e.mesh,service:e.name}},v={name:"mesh-detail-view",params:{mesh:e.mesh}};let a="\u2014";if(e.dataplanes){const{online:w=0,total:k=0}=e.dataplanes;a=`${w} / ${k}`}let u="\u2014";e.status&&(u=P[e.status].title);const s=(r=e.serviceType)!=null?r:"internal";return{...e,serviceType:s,nameRoute:i,meshRoute:v,dpProxiesStatus:a,status:u}}function y(e){t.value=e}return(e,i)=>(S(),f(L,null,{content:b(()=>[g(M,{"page-size":d,error:l.value,"has-error":l.value!==null,"is-loading":c.value,"empty-state":x,"table-data":o.value,"table-data-is-empty":o.value.data.length===0,next:h.value,onTableAction:y,onLoadData:p},null,8,["error","has-error","is-loading","table-data","table-data-is-empty","next"])]),sidebar:b(()=>[t.value!==null?(S(),f(N,{key:0,name:t.value.name,mesh:t.value.mesh,"service-type":t.value.serviceType},null,8,["name","mesh","service-type"])):z("",!0)]),_:1}))}});export{H as default};
