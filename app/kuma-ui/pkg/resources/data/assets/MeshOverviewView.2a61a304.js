import{d as O,f as u,g as F,o as l,j as i,a as h,u as t,e as c,q as X,E as K,p as Q,r as C,y as U,cs as R,k as S,c as I,w as f,l as d,F as b,n as B,t as m,Q as $,A as N}from"./index.60b0f0ac.js";import{_ as L,a as T,M as W}from"./MeshResources.0afb8bc8.js";import{_ as A}from"./LabelList.vue_vue_type_style_index_0_lang.bd2c37a0.js";import{T as z}from"./TabsWidget.5b63a728.js";import{_ as G}from"./YamlView.vue_vue_type_script_setup_true_lang.152633f3.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.548da37c.js";import"./ErrorBlock.2ee4d08e.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.be7e4bb1.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.85e36160.js";import"./_commonjsHelpers.f037b798.js";const H={class:"chart-container mt-16"},J=O({__name:"MeshCharts",setup(Y){const s=X(),y=u(()=>s.getters.getServiceResourcesFetching),M=u(()=>s.getters.getMeshInsightsFetching),p=u(()=>s.getters.getChart("services")),g=u(()=>s.getters.getChart("dataplanes")),_=u(()=>s.getters.getChart("kumaDPVersions")),o=u(()=>s.getters.getChart("envoyVersions"));F(()=>s.state.selectedMesh,function(){w()}),w();function w(){s.dispatch("fetchMeshInsights",s.state.selectedMesh),s.dispatch("fetchServices",s.state.selectedMesh)}return(E,V)=>(l(),i("div",H,[h(L,{class:"chart chart-1/4",title:{singular:"SERVICE",plural:"SERVICES"},data:t(p).data,"is-loading":t(y),"save-chart":""},null,8,["data","is-loading"]),c(),h(L,{class:"chart chart-1/4",title:{singular:"DP PROXY",plural:"DP PROXIES"},data:t(g).data,url:{name:"data-plane-list-view",params:{mesh:t(s).state.selectedMesh}},"is-loading":t(M)},null,8,["data","url","is-loading"]),c(),h(T,{class:"chart chart-1/4",title:"KUMA DP",data:t(_).data,"is-loading":t(M)},null,8,["data","is-loading"]),c(),h(T,{class:"chart chart-1/4",title:"ENVOY",data:t(o).data,"is-loading":t(M),"display-am-charts-logo":""},null,8,["data","is-loading"])]))}});const Z=K(J,[["__scopeId","data-v-da78099c"]]),ee={key:1},ae={key:1},te={key:1,class:"mt-8"},pe=O({__name:"MeshOverviewView",setup(Y){const s=Q(),y=X(),M=[{hash:"#overview",title:"Overview"},{hash:"#resources",title:"Resources"}],p=C(!0),g=C(!1),_=C(!1),o=C(null),w=C(null),E=u(()=>o.value!==null?U(o.value):null),V=u(()=>{if(o.value===null)return null;const{name:n,type:r,creationTime:e,modificationTime:a}=o.value;return{name:n,type:r,created:R(e),modified:R(a)}}),j=u(()=>{var k;if(o.value===null)return null;const n=D(o.value,"mtls"),r=D(o.value,"logging"),e=D(o.value,"metrics"),a=D(o.value,"tracing"),v=Boolean((k=o.value.routing)==null?void 0:k.localityAwareLoadBalancing);return{mtls:n,logging:r,metrics:e,tracing:a,localityAwareLoadBalancing:v}}),x=u(()=>{const n=y.state.policies.map(r=>{var e,a;return{title:r.pluralDisplayName,value:(a=(e=y.state.meshInsight.policies[r.name])==null?void 0:e.total)!=null?a:0}});return[{title:"Data Plane Proxies",value:y.state.meshInsight.dataplanes.total},...n]});F(()=>s.params.mesh,function(){s.name==="single-mesh-overview"&&(p.value=!0,_.value=!1,g.value=!1,P())}),P();async function P(){p.value=!0,_.value=!1;const n=s.params.mesh;try{o.value=await S.getMesh({name:n}),w.value=await S.getMeshInsights({name:n})}catch(r){g.value=!0,_.value=!0,console.error(r)}finally{p.value=!1}}function D(n,r){var v,k;if(n===null||n[r]===void 0)return!1;const e=(k=(v=n[r].enabledBackend)!=null?v:n[r].defaultBackend)!=null?k:n[r].backends[0].name,a=n[r].backends.find(q=>q.name===e);return`${a.type} / ${a.name}`}return(n,r)=>(l(),i(b,null,[h(Z),c(),h(W,{class:"mt-8"}),c(),o.value!==null?(l(),I(z,{key:0,class:"mt-8","has-error":g.value,"is-loading":p.value,tabs:M,"initial-tab-override":"overview"},{overview:f(()=>[h(A,null,{default:f(()=>[d("div",null,[d("ul",null,[(l(!0),i(b,null,B(t(V),(e,a)=>(l(),i("li",{key:a},[d("h4",null,m(a),1),c(),typeof e=="boolean"?(l(),I(t($),{key:0,appearance:e?"success":"danger"},{default:f(()=>[c(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),i("p",ee,m(e),1))]))),128))])]),c(),d("div",null,[d("ul",null,[(l(!0),i(b,null,B(t(j),(e,a)=>(l(),i("li",{key:a},[d("h4",null,m(a),1),c(),typeof e=="boolean"?(l(),I(t($),{key:0,appearance:e?"success":"danger"},{default:f(()=>[c(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),i("p",ae,m(e),1))]))),128))])])]),_:1})]),resources:f(()=>[h(A,{"has-error":g.value,"is-loading":p.value,"is-empty":_.value},{default:f(()=>[(l(!0),i(b,null,B(Math.ceil(t(x).length/3),e=>(l(),i("div",{key:e},[d("ul",null,[(l(!0),i(b,null,B(t(x).slice((e-1)*3,e*3),(a,v)=>(l(),i("li",{key:v},[d("h4",null,m(a.title),1),c(),d("p",null,m(a.value),1)]))),128))])]))),128))]),_:1},8,["has-error","is-loading","is-empty"])]),_:1},8,["has-error","is-loading"])):N("",!0),c(),t(E)!==null?(l(),i("div",te,[h(G,{id:"code-block-mesh",content:t(E)},null,8,["content"])])):N("",!0)],64))}});export{pe as default};
