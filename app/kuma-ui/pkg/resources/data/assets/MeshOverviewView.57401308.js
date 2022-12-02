import{d as A,e as c,f as F,o as l,i,a as d,u as t,q as Y,D as z,p as K,r as k,x as U,cs as S,k as P,c as B,w as v,j as u,F as b,n as D,t as h,S as R,b as N,z as $}from"./index.3bc39668.js";import{_ as L,a as T,M as W}from"./MeshResources.d7c8256d.js";import{_ as O}from"./LabelList.vue_vue_type_style_index_0_lang.0e14ac31.js";import{T as G}from"./TabsWidget.1751eed8.js";import{Y as H}from"./YamlView.24c9d3cb.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.74b6b406.js";import"./ErrorBlock.f4ac98cc.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.13b03cfc.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.b3d38a49.js";import"./_commonjsHelpers.f037b798.js";const J={class:"chart-container mt-16"},Q=A({__name:"MeshCharts",setup(X){const s=Y(),f=c(()=>s.getters.getServiceResourcesFetching),y=c(()=>s.getters.getMeshInsightsFetching),m=c(()=>s.getters.getChart("services")),p=c(()=>s.getters.getChart("dataplanes")),g=c(()=>s.getters.getChart("kumaDPVersions")),o=c(()=>s.getters.getChart("envoyVersions"));F(()=>s.state.selectedMesh,function(){C()}),C();function C(){s.dispatch("fetchMeshInsights",s.state.selectedMesh),s.dispatch("fetchServices",s.state.selectedMesh)}return(V,x)=>(l(),i("div",J,[d(L,{class:"chart chart-1/4",title:{singular:"SERVICE",plural:"SERVICES"},data:t(m).data,"is-loading":t(f),"save-chart":""},null,8,["data","is-loading"]),d(L,{class:"chart chart-1/4",title:{singular:"DP PROXY",plural:"DP PROXIES"},data:t(p).data,url:{name:"data-plane-list-view",params:{mesh:t(s).state.selectedMesh}},"is-loading":t(y)},null,8,["data","url","is-loading"]),d(T,{class:"chart chart-1/4",title:"KUMA DP",data:t(g).data,"is-loading":t(y)},null,8,["data","is-loading"]),d(T,{class:"chart chart-1/4",title:"ENVOY",data:t(o).data,"is-loading":t(y),"display-am-charts-logo":""},null,8,["data","is-loading"])]))}});const Z=z(Q,[["__scopeId","data-v-da78099c"]]),ee={key:1},ae={key:1},te={key:1,class:"mt-8"},pe=A({__name:"MeshOverviewView",setup(X){const s=K(),f=Y(),y=[{hash:"#overview",title:"Overview"},{hash:"#resources",title:"Resources"}],m=k(!0),p=k(!1),g=k(!1),o=k(null),C=k(null),V=c(()=>o.value!==null?U(o.value):null),x=c(()=>{if(o.value===null)return null;const{name:n,type:r,creationTime:e,modificationTime:a}=o.value;return{name:n,type:r,created:S(e),modified:S(a)}}),j=c(()=>{var M;if(o.value===null)return null;const n=w(o.value,"mtls"),r=w(o.value,"logging"),e=w(o.value,"metrics"),a=w(o.value,"tracing"),_=Boolean((M=o.value.routing)==null?void 0:M.localityAwareLoadBalancing);return{mtls:n,logging:r,metrics:e,tracing:a,localityAwareLoadBalancing:_}}),E=c(()=>{const n=f.state.policies.map(r=>{var e,a;return{title:r.pluralDisplayName,value:(a=(e=f.state.meshInsight.policies[r.name])==null?void 0:e.total)!=null?a:0}});return[{title:"Data Plane Proxies",value:f.state.meshInsight.dataplanes.total},...n]});F(()=>s.params.mesh,function(){s.name==="single-mesh-overview"&&(m.value=!0,g.value=!1,p.value=!1,I())}),I();async function I(){m.value=!0,g.value=!1;const n=s.params.mesh;try{o.value=await P.getMesh({name:n}),C.value=await P.getMeshInsights({name:n})}catch(r){p.value=!0,g.value=!0,console.error(r)}finally{m.value=!1}}function w(n,r){var _,M;if(n===null||n[r]===void 0)return!1;const e=(M=(_=n[r].enabledBackend)!=null?_:n[r].defaultBackend)!=null?M:n[r].backends[0].name,a=n[r].backends.find(q=>q.name===e);return`${a.type} / ${a.name}`}return(n,r)=>(l(),i(b,null,[d(Z),d(W,{class:"mt-8"}),o.value!==null?(l(),B(G,{key:0,class:"mt-8","has-error":p.value,"is-loading":m.value,tabs:y,"initial-tab-override":"overview"},{overview:v(()=>[d(O,null,{default:v(()=>[u("div",null,[u("ul",null,[(l(!0),i(b,null,D(t(x),(e,a)=>(l(),i("li",{key:a},[u("h4",null,h(a),1),typeof e=="boolean"?(l(),B(t(R),{key:0,appearance:e?"success":"danger"},{default:v(()=>[N(h(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),i("p",ee,h(e),1))]))),128))])]),u("div",null,[u("ul",null,[(l(!0),i(b,null,D(t(j),(e,a)=>(l(),i("li",{key:a},[u("h4",null,h(a),1),typeof e=="boolean"?(l(),B(t(R),{key:0,appearance:e?"success":"danger"},{default:v(()=>[N(h(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),i("p",ae,h(e),1))]))),128))])])]),_:1})]),resources:v(()=>[d(O,{"has-error":p.value,"is-loading":m.value,"is-empty":g.value},{default:v(()=>[(l(!0),i(b,null,D(Math.ceil(t(E).length/3),e=>(l(),i("div",{key:e},[u("ul",null,[(l(!0),i(b,null,D(t(E).slice((e-1)*3,e*3),(a,_)=>(l(),i("li",{key:_},[u("h4",null,h(a.title),1),u("p",null,h(a.value),1)]))),128))])]))),128))]),_:1},8,["has-error","is-loading","is-empty"])]),_:1},8,["has-error","is-loading"])):$("",!0),t(V)!==null?(l(),i("div",te,[d(H,{id:"code-block-mesh",content:t(V)},null,8,["content"])])):$("",!0)],64))}});export{pe as default};
