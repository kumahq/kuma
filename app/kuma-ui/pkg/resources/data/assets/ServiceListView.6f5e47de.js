import{d as A,p as D,r as n,g as C,k as g,cC as E,o as h,c as b,w as S,a as P,y as L}from"./index.438e3d4b.js";import{C as M}from"./ContentWrapper.d992a5a6.js";import{D as N}from"./DataOverview.d3cf1e01.js";import{S as R}from"./ServiceDetails.c2dc24c2.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.5a7a3b48.js";import"./ErrorBlock.aea25275.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.5397142f.js";import"./TagList.9c7242fd.js";import"./EntityURLControl.vue_vue_type_script_setup_true_lang.9cb692f4.js";import"./YamlView.5dc1bbe6.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.b3ce259f.js";import"./_commonjsHelpers.f037b798.js";const J=A({__name:"ServiceListView",setup(V){const T=[{label:"Service",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"serviceType"},{label:"Address",key:"address"},{label:"Status",key:"status"},{label:"DP proxies (online / total)",key:"dpProxiesStatus"}],d=50,x={title:"No Data",message:"There are no service insights present."},u=D(),c=n(!0),p=n(null),y=n(null),t=n(null),o=n({headers:T,data:[]});C(()=>u.params.mesh,function(){u.name==="service-list-view"&&v(0)}),v(0);async function v(e){c.value=!0,p.value=null;const i=u.params.mesh,l=d;try{const{items:a=[],next:m}=await g.getAllServiceInsightsFromMesh({mesh:i},{size:l,offset:e});y.value=m,Array.isArray(a)&&a.length>0?(a.sort((s,r)=>s.name>r.name?1:s.name<r.name?-1:s.mesh.localeCompare(r.mesh)),f(a[0]),o.value.data=a.map(s=>k(s))):(t.value=null,o.value.data=[])}catch(a){t.value=null,a instanceof Error?p.value=a:console.error(a)}finally{c.value=!1}}function k(e){var r;const i={name:e.serviceType==="external"?"external-service-detail-view":"service-insight-detail-view",params:{mesh:e.mesh,service:e.name}},l={name:"mesh-detail-view",params:{mesh:e.mesh}};let a="\u2014";if(e.dataplanes){const{online:_=0,total:w=0}=e.dataplanes;a=`${_} / ${w}`}let m="\u2014";e.status&&(m=E[e.status].title);const s=(r=e.serviceType)!=null?r:"internal";return{...e,serviceType:s,nameRoute:i,meshRoute:l,dpProxiesStatus:a,status:m}}function f(e){t.value=e}return(e,i)=>(h(),b(M,null,{content:S(()=>{var l;return[P(N,{"selected-entity-name":(l=t.value)==null?void 0:l.name,"page-size":d,error:p.value,"is-loading":c.value,"empty-state":x,"table-data":o.value,"table-data-is-empty":o.value.data.length===0,next:y.value,onTableAction:f,onLoadData:v},null,8,["selected-entity-name","error","is-loading","table-data","table-data-is-empty","next"])]}),sidebar:S(()=>[t.value!==null?(h(),b(R,{key:0,name:t.value.name,mesh:t.value.mesh,"service-type":t.value.serviceType},null,8,["name","mesh","service-type"])):L("",!0)]),_:1}))}});export{J as default};
