import{F as x,A}from"./kongponents.es-a62ec79b.js";import{d as N,c as m,v as S,o as n,j as d,g as r,b as o,h as i,u as G,r as C,K as H,L as T,x as J,w as l,e as p,f as B,i as E,F as f,q as V,t as v}from"./index-5550fdb4.js";import{D as w}from"./DoughnutChart-eda28231.js";import{u as F}from"./store-72e7410d.js";import{_ as K}from"./_plugin-vue_export-helper-c27b6911.js";import{a as P,D as I}from"./DefinitionListItem-5764f62f.js";import{M as Q}from"./MeshResources-fc83098d.js";import{_ as U}from"./StatusInfo.vue_vue_type_script_setup_true_lang-cd8522cf.js";import{_ as W}from"./YamlView.vue_vue_type_script_setup_true_lang-f5b58b07.js";import{u as X}from"./index-6d226453.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-98379006.js";import"./ErrorBlock-396acb12.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-61797f64.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-caf13288.js";import"./toYaml-4e00099e.js";const Y={class:"chart-box-list"},Z=N({__name:"MeshCharts",setup(O){const c=F(),g=m(()=>c.getters.getChart("services",{title:"Services",showTotal:!0})),h=m(()=>c.getters.getChart("dataplanes",{title:"DP Proxies",showTotal:!0,isStatusChart:!0})),y=m(()=>c.getters.getChart("kumaDPVersions",{title:"Kuma DP",subtitle:"versions"})),k=m(()=>c.getters.getChart("envoyVersions",{title:"Envoy",subtitle:"versions"}));S(()=>c.state.selectedMesh,function(){t()}),t();function t(){c.dispatch("fetchMeshInsights",c.state.selectedMesh),c.dispatch("fetchServices",c.state.selectedMesh)}return(b,M)=>(n(),d("div",Y,[r(w,{data:o(g)},null,8,["data"]),i(),r(w,{data:o(h)},null,8,["data"]),i(),r(w,{data:o(y)},null,8,["data"]),i(),r(w,{data:o(k)},null,8,["data"])]))}});const ee=K(Z,[["__scopeId","data-v-7d682009"]]),te={class:"kcard-stack"},ae={class:"columns"},se={key:0},ne=N({__name:"MeshOverviewView",setup(O){const c=X(),g=G(),h=F(),y=C(!0),k=C(!1),t=C(null),b=C(null),M=m(()=>t.value!==null?H(t.value):null),R=m(()=>{if(t.value===null)return null;const{name:a,type:s,creationTime:_,modificationTime:e}=t.value;return{name:a,type:s,created:T(_),modified:T(e),"Data Plane Proxies":h.state.meshInsight.dataplanes.total}}),j=m(()=>{var $;if(t.value===null)return null;const a=D(t.value,"mtls"),s=D(t.value,"logging"),_=D(t.value,"metrics"),e=D(t.value,"tracing"),u=!!(($=t.value.routing)!=null&&$.localityAwareLoadBalancing);return{mtls:a,logging:s,metrics:_,tracing:e,localityAwareLoadBalancing:u}}),q=m(()=>h.state.sidebar.insights.mesh.policies.total),z=m(()=>h.state.policyTypes.map(a=>{var s;return{...a,length:((s=h.state.meshInsight.policies[a.name])==null?void 0:s.total)??0}}));S(()=>g.params.mesh,function(){g.name==="single-mesh-overview"&&L()}),L();async function L(){y.value=!0,k.value=!1;const a=g.params.mesh;try{t.value=await c.getMesh({name:a}),b.value=await c.getMeshInsights({name:a})}catch(s){k.value=!0,t.value=null,b.value=null,console.error(s)}finally{y.value=!1}}function D(a,s){if(a===null||a[s]===void 0)return!1;const _=a[s].enabledBackend??a[s].defaultBackend??a[s].backends[0].name,e=a[s].backends.find(u=>u.name===_);return`${e.type} / ${e.name}`}return(a,s)=>{const _=J("router-link");return n(),d("div",te,[r(o(x),null,{body:l(()=>[r(ee)]),_:1}),i(),t.value!==null?(n(),p(o(x),{key:0},{body:l(()=>[E("div",ae,[r(U,{"is-loading":y.value,"has-error":k.value,"is-empty":t.value===null||b.value===null},{default:l(()=>[r(P,null,{default:l(()=>[(n(!0),d(f,null,V(o(R),(e,u)=>(n(),p(I,{key:u,term:u},{default:l(()=>[typeof e=="boolean"?(n(),p(o(A),{key:0,appearance:e?"success":"danger"},{default:l(()=>[i(v(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),d(f,{key:1},[i(v(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),_:1},8,["is-loading","has-error","is-empty"]),i(),r(P,null,{default:l(()=>[(n(!0),d(f,null,V(o(j),(e,u)=>(n(),p(I,{key:u,term:u},{default:l(()=>[typeof e=="boolean"?(n(),p(o(A),{key:0,appearance:e?"success":"danger"},{default:l(()=>[i(v(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),d(f,{key:1},[i(v(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1}),i(),r(P,null,{default:l(()=>[r(I,{term:`Policies (${o(q)})`},{default:l(()=>[E("ul",null,[(n(!0),d(f,null,V(o(z),(e,u)=>(n(),d(f,{key:u},[e.length!==0?(n(),d("li",se,[r(_,{to:{name:"policy-list-view",params:{policyPath:e.path}}},{default:l(()=>[i(v(e.name)+": "+v(e.length),1)]),_:2},1032,["to"])])):B("",!0)],64))),128))])]),_:1},8,["term"])]),_:1})])]),_:1})):B("",!0),i(),o(M)!==null?(n(),p(o(x),{key:1},{body:l(()=>[r(W,{id:"code-block-mesh",content:o(M)},null,8,["content"])]),_:1})):B("",!0),i(),r(Q)])}}});const ke=K(ne,[["__scopeId","data-v-4d430812"]]);export{ke as default};
