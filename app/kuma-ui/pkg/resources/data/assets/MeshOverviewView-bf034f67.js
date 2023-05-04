import{F as x,x as L}from"./kongponents.es-ba82ceca.js";import{d as S,c as m,v as A,o as n,j as d,g as r,b as o,h as i,u as z,r as C,E as H,G as T,x as J,w as l,e as p,f as B,i as N,F as f,q as V,t as v}from"./index-bd38c154.js";import{D as w}from"./DoughnutChart-5b4294aa.js";import{u as F}from"./store-82b3ee45.js";import{_ as O}from"./_plugin-vue_export-helper-c27b6911.js";import{a as P,D as I}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-d6e052e1.js";import{M as Q}from"./MeshResources-1d58ac74.js";import{_ as U}from"./StatusInfo.vue_vue_type_script_setup_true_lang-12b7659b.js";import{_ as W}from"./YamlView.vue_vue_type_script_setup_true_lang-b71e2184.js";import{u as X}from"./index-2e645818.js";import"./ErrorBlock-99d9a9e3.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-fc8e76df.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-39ef11a2.js";import"./toYaml-4e00099e.js";const Y={class:"chart-box-list"},Z=S({__name:"MeshCharts",setup(R){const c=F(),g=m(()=>c.getters.getChart("services",{title:"Services",showTotal:!0})),h=m(()=>c.getters.getChart("dataplanes",{title:"DP Proxies",showTotal:!0,isStatusChart:!0})),y=m(()=>c.getters.getChart("kumaDPVersions",{title:"Kuma DP",subtitle:"versions"})),k=m(()=>c.getters.getChart("envoyVersions",{title:"Envoy",subtitle:"versions"}));A(()=>c.state.selectedMesh,function(){t()}),t();function t(){c.dispatch("fetchMeshInsights",c.state.selectedMesh),c.dispatch("fetchServices",c.state.selectedMesh)}return(b,M)=>(n(),d("div",Y,[r(w,{data:o(g)},null,8,["data"]),i(),r(w,{data:o(h)},null,8,["data"]),i(),r(w,{data:o(y)},null,8,["data"]),i(),r(w,{data:o(k)},null,8,["data"])]))}});const ee=O(Z,[["__scopeId","data-v-7d682009"]]),te={class:"kcard-stack"},ae={class:"columns"},se={key:0},ne=S({__name:"MeshOverviewView",setup(R){const c=X(),g=z(),h=F(),y=C(!0),k=C(!1),t=C(null),b=C(null),M=m(()=>t.value!==null?H(t.value):null),K=m(()=>{if(t.value===null)return null;const{name:a,type:s,creationTime:_,modificationTime:e}=t.value;return{name:a,type:s,created:T(_),modified:T(e),"Data Plane Proxies":h.state.meshInsight.dataplanes.total}}),j=m(()=>{var E;if(t.value===null)return null;const a=D(t.value,"mtls"),s=D(t.value,"logging"),_=D(t.value,"metrics"),e=D(t.value,"tracing"),u=!!((E=t.value.routing)!=null&&E.localityAwareLoadBalancing);return{mtls:a,logging:s,metrics:_,tracing:e,localityAwareLoadBalancing:u}}),q=m(()=>h.state.sidebar.insights.mesh.policies.total),G=m(()=>h.state.policyTypes.map(a=>{var s;return{...a,length:((s=h.state.meshInsight.policies[a.name])==null?void 0:s.total)??0}}));A(()=>g.params.mesh,function(){g.name==="single-mesh-overview"&&$()}),$();async function $(){y.value=!0,k.value=!1;const a=g.params.mesh;try{t.value=await c.getMesh({name:a}),b.value=await c.getMeshInsights({name:a})}catch(s){k.value=!0,t.value=null,b.value=null,console.error(s)}finally{y.value=!1}}function D(a,s){if(a===null||a[s]===void 0)return!1;const _=a[s].enabledBackend??a[s].defaultBackend??a[s].backends[0].name,e=a[s].backends.find(u=>u.name===_);return`${e.type} / ${e.name}`}return(a,s)=>{const _=J("router-link");return n(),d("div",te,[r(o(x),null,{body:l(()=>[r(ee)]),_:1}),i(),t.value!==null?(n(),p(o(x),{key:0},{body:l(()=>[N("div",ae,[r(U,{"is-loading":y.value,"has-error":k.value,"is-empty":t.value===null||b.value===null},{default:l(()=>[r(P,null,{default:l(()=>[(n(!0),d(f,null,V(o(K),(e,u)=>(n(),p(I,{key:u,term:u},{default:l(()=>[typeof e=="boolean"?(n(),p(o(L),{key:0,appearance:e?"success":"danger"},{default:l(()=>[i(v(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),d(f,{key:1},[i(v(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),_:1},8,["is-loading","has-error","is-empty"]),i(),r(P,null,{default:l(()=>[(n(!0),d(f,null,V(o(j),(e,u)=>(n(),p(I,{key:u,term:u},{default:l(()=>[typeof e=="boolean"?(n(),p(o(L),{key:0,appearance:e?"success":"danger"},{default:l(()=>[i(v(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),d(f,{key:1},[i(v(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1}),i(),r(P,null,{default:l(()=>[r(I,{term:`Policies (${o(q)})`},{default:l(()=>[N("ul",null,[(n(!0),d(f,null,V(o(G),(e,u)=>(n(),d(f,{key:u},[e.length!==0?(n(),d("li",se,[r(_,{to:{name:"policy-list-view",params:{policyPath:e.path}}},{default:l(()=>[i(v(e.name)+": "+v(e.length),1)]),_:2},1032,["to"])])):B("",!0)],64))),128))])]),_:1},8,["term"])]),_:1})])]),_:1})):B("",!0),i(),o(M)!==null?(n(),p(o(x),{key:1},{body:l(()=>[r(W,{id:"code-block-mesh",content:o(M)},null,8,["content"])]),_:1})):B("",!0),i(),r(Q)])}}});const ye=O(ne,[["__scopeId","data-v-4d430812"]]);export{ye as default};
