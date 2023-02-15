import{u as j}from"./vue-router-cf3250ac.js";import{c as I}from"./kongponents.es-8abed680.js";import{C as q}from"./ContentWrapper-856507ff.js";import{u as N}from"./store-0511bcbf.js";import{D as M}from"./DoughnutChart-ffc86670.js";import{d as S,c as p,s as L,o as n,h as r,e as d,u as c,f as o,r as y,i as z,w as k,b as D,F as b,g as u,m as x,t as m,a as E}from"./runtime-dom.esm-bundler-a6f4ece5.js";import{_ as A}from"./_plugin-vue_export-helper-c27b6911.js";import{M as G}from"./MeshResources-789dc1be.js";import{_ as H}from"./LabelList.vue_vue_type_style_index_0_lang-5b409b55.js";import{_ as J}from"./YamlView.vue_vue_type_script_setup_true_lang-7c577dfe.js";import{k as T}from"./kumaApi-de9fdcba.js";import{o as Q,k as $}from"./production-0f1ffdb6.js";import"./index-28f79c9b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-c8983716.js";import"./ErrorBlock-31dfb839.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-55d18240.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-be31e96b.js";import"./toYaml-4e00099e.js";const U={class:"chart-box-list"},X=S({__name:"MeshCharts",setup(O){const l=N(),_=p(()=>l.getters.getChart("services",{title:"Services",showTotal:!0})),f=p(()=>l.getters.getChart("dataplanes",{title:"DP Proxies",showTotal:!0,isStatusChart:!0})),g=p(()=>l.getters.getChart("kumaDPVersions",{title:"Kuma DP",subtitle:"versions"})),v=p(()=>l.getters.getChart("envoyVersions",{title:"Envoy",subtitle:"versions"}));L(()=>l.state.selectedMesh,function(){a()}),a();function a(){l.dispatch("fetchMeshInsights",l.state.selectedMesh),l.dispatch("fetchServices",l.state.selectedMesh)}return(B,w)=>(n(),r("div",U,[d(M,{data:c(_)},null,8,["data"]),o(),d(M,{data:c(f)},null,8,["data"]),o(),d(M,{data:c(g)},null,8,["data"]),o(),d(M,{data:c(v)},null,8,["data"])]))}});const Y=A(X,[["__scopeId","data-v-62ff2edc"]]),Z={key:0},ee={key:1},te={key:1},ae={class:"policy-counts"},se={key:0},ne={key:0,class:"mt-4"},oe=S({__name:"MeshOverviewView",setup(O){const l=j(),_=N(),f=y(!0),g=y(!1),v=y(!1),a=y(null),B=y(null),w=p(()=>a.value!==null?Q(a.value):null),R=p(()=>{if(a.value===null)return null;const{name:t,type:s,creationTime:h,modificationTime:e}=a.value;return{name:t,type:s,created:$(h),modified:$(e),"Data Plane Proxies":_.state.meshInsight.dataplanes.total}}),F=p(()=>{var P;if(a.value===null)return null;const t=C(a.value,"mtls"),s=C(a.value,"logging"),h=C(a.value,"metrics"),e=C(a.value,"tracing"),i=Boolean((P=a.value.routing)==null?void 0:P.localityAwareLoadBalancing);return{mtls:t,logging:s,metrics:h,tracing:e,localityAwareLoadBalancing:i}}),K=p(()=>_.state.sidebar.insights.mesh.policies.total),W=p(()=>_.state.policyTypes.map(t=>{var s;return{...t,length:((s=_.state.meshInsight.policies[t.name])==null?void 0:s.total)??0}}));L(()=>l.params.mesh,function(){l.name==="single-mesh-overview"&&(f.value=!0,v.value=!1,g.value=!1,V())}),V();async function V(){f.value=!0,v.value=!1;const t=l.params.mesh;try{a.value=await T.getMesh({name:t}),B.value=await T.getMeshInsights({name:t})}catch(s){g.value=!0,v.value=!0,console.error(s)}finally{f.value=!1}}function C(t,s){if(t===null||t[s]===void 0)return!1;const h=t[s].enabledBackend??t[s].defaultBackend??t[s].backends[0].name,e=t[s].backends.find(i=>i.name===h);return`${e.type} / ${e.name}`}return(t,s)=>{const h=z("router-link");return n(),r(b,null,[d(Y,{class:"mt-24"}),o(),d(q,{class:"mt-8"},{content:k(()=>[a.value!==null?(n(),r("div",Z,[d(H,{"has-error":g.value,"is-loading":f.value,"is-empty":v.value},{default:k(()=>[u("div",null,[u("ul",null,[(n(!0),r(b,null,x(c(R),(e,i)=>(n(),r("li",{key:i},[u("h4",null,m(i),1),o(),typeof e=="boolean"?(n(),E(c(I),{key:0,appearance:e?"success":"danger"},{default:k(()=>[o(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),r("p",ee,m(e),1))]))),128))])]),o(),u("div",null,[u("ul",null,[(n(!0),r(b,null,x(c(F),(e,i)=>(n(),r("li",{key:i},[u("h4",null,m(i),1),o(),typeof e=="boolean"?(n(),E(c(I),{key:0,appearance:e?"success":"danger"},{default:k(()=>[o(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(n(),r("p",te,m(e),1))]))),128))])]),o(),u("div",null,[u("ul",ae,[u("li",null,[u("h4",null,`
                  Policies (`+m(c(K))+`)
                `,1),o(),u("ul",null,[(n(!0),r(b,null,x(c(W),(e,i)=>(n(),r(b,{key:i},[e.length!==0?(n(),r("li",se,[d(h,{to:{name:"policy",params:{policyPath:e.path}}},{default:k(()=>[o(m(e.name)+": "+m(e.length),1)]),_:2},1032,["to"])])):D("",!0)],64))),128))])])])])]),_:1},8,["has-error","is-loading","is-empty"])])):D("",!0)]),_:1}),o(),c(w)!==null?(n(),r("div",ne,[d(J,{id:"code-block-mesh",content:c(w)},null,8,["content"])])):D("",!0),o(),d(G,{class:"mt-6"})],64)}}});const we=A(oe,[["__scopeId","data-v-1ccb92a3"]]);export{we as default};
