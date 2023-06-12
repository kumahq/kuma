import{Z as x,P as E}from"./kongponents.es-7e228e6a.js";import{d as A,q as O,m as _,s as R,o as r,f as d,a as s,b as l,r as D,M as N,v as G,c as p,w as o,u as m,e as B,k as P,t as g,F as y,g as S}from"./index-2aa994fe.js";import{b as F,D as M,_ as K,g as H,u as J,f as Q,e as U}from"./RouteView.vue_vue_type_script_setup_true_lang-49cffe7a.js";import{_ as X}from"./RouteTitle.vue_vue_type_script_setup_true_lang-994a27b4.js";import{D as $,a as V}from"./DefinitionListItem-7bd1502c.js";import{M as Y}from"./MeshResources-e58deef6.js";import{_ as ee}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-0bfb6b79.js";import{_ as te}from"./StatusInfo.vue_vue_type_script_setup_true_lang-596f3434.js";import{T as ae}from"./TextWithCopyButton-71e1d5a6.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-55ca759f.js";import"./toYaml-4e00099e.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-d8e16f78.js";import"./ErrorBlock-9a3c452b.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-c25833c3.js";const se={class:"chart-box-list"},ne=A({__name:"MeshCharts",setup(Z){const k=O(),u=F(),f=_(()=>u.getters.getChart("services",{title:"Services",showTotal:!0})),v=_(()=>u.getters.getChart("dataplanes",{title:"DP Proxies",showTotal:!0,isStatusChart:!0})),b=_(()=>u.getters.getChart("kumaDPVersions",{title:"Kuma DP",subtitle:"versions"})),w=_(()=>u.getters.getChart("envoyVersions",{title:"Envoy",subtitle:"versions"}));R(()=>k.params.mesh,function(c){typeof c=="string"&&n(c)},{immediate:!0});function n(c){u.dispatch("fetchMeshInsights",c),u.dispatch("fetchServices",c)}return(c,I)=>(r(),d("div",se,[s(M,{data:f.value},null,8,["data"]),l(),s(M,{data:v.value},null,8,["data"]),l(),s(M,{data:b.value},null,8,["data"]),l(),s(M,{data:w.value},null,8,["data"])]))}});const oe=K(ne,[["__scopeId","data-v-5ab6d374"]]),re={class:"kcard-stack"},le={class:"columns"},ie={key:0},ue=A({__name:"MeshOverviewView",setup(Z){const{t:k}=H(),u=J(),f=O(),v=F(),b=D(!0),w=D(!1),n=D(null),c=D(null),I=_(()=>{if(n.value===null)return null;const{name:t,creationTime:a,modificationTime:h}=n.value;return{name:t,created:N(a),modified:N(h),"Data Plane Proxies":v.state.meshInsight.dataplanes.total}}),q=_(()=>{var L;if(n.value===null)return null;const t=C(n.value,"mtls"),a=C(n.value,"logging"),h=C(n.value,"metrics"),e=C(n.value,"tracing"),i=!!((L=n.value.routing)!=null&&L.localityAwareLoadBalancing);return{mtls:t,logging:a,metrics:h,tracing:e,localityAwareLoadBalancing:i}}),W=_(()=>v.state.sidebar.insights.mesh.policies.total),j=_(()=>v.state.policyTypes.map(t=>{var a;return{...t,length:((a=v.state.meshInsight.policies[t.name])==null?void 0:a.total)??0}}));R(()=>f.params.mesh,function(){f.name==="single-mesh-overview"&&T()}),T();async function T(){b.value=!0,w.value=!1;const t=f.params.mesh;try{n.value=await u.getMesh({name:t}),c.value=await u.getMeshInsights({name:t})}catch(a){w.value=!0,n.value=null,c.value=null,console.error(a)}finally{b.value=!1}}function C(t,a){if(t===null||t[a]===void 0)return!1;const h=t[a].enabledBackend??t[a].defaultBackend??t[a].backends[0].name,e=t[a].backends.find(i=>i.name===h);return`${e.type} / ${e.name}`}async function z(t){const a=f.params.mesh;return await u.getMesh({name:a},t)}return(t,a)=>{const h=G("router-link");return r(),p(U,null,{default:o(()=>[s(X,{title:m(k)("meshes.routes.overview.title")},null,8,["title"]),l(),s(Q,null,{default:o(()=>[B("div",re,[s(m(x),null,{body:o(()=>[s(oe)]),_:1}),l(),n.value!==null?(r(),p(m(x),{key:0},{body:o(()=>[B("div",le,[s(te,{"is-loading":b.value,"has-error":w.value,"is-empty":n.value===null||c.value===null},{default:o(()=>[s($,null,{default:o(()=>[(r(!0),d(y,null,P(I.value,(e,i)=>(r(),p(V,{key:i,term:m(k)(`http.api.property.${i}`)},{default:o(()=>[typeof e=="boolean"?(r(),p(m(E),{key:0,appearance:e?"success":"danger"},{default:o(()=>[l(g(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):i==="name"&&typeof e=="string"?(r(),p(ae,{key:1,text:e},null,8,["text"])):(r(),d(y,{key:2},[l(g(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),_:1},8,["is-loading","has-error","is-empty"]),l(),s($,null,{default:o(()=>[(r(!0),d(y,null,P(q.value,(e,i)=>(r(),p(V,{key:i,term:m(k)(`http.api.property.${i}`)},{default:o(()=>[typeof e=="boolean"?(r(),p(m(E),{key:0,appearance:e?"success":"danger"},{default:o(()=>[l(g(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(r(),d(y,{key:1},[l(g(e),1)],64))]),_:2},1032,["term"]))),128))]),_:1}),l(),s($,null,{default:o(()=>[s(V,{term:`Policies (${W.value})`},{default:o(()=>[B("ul",null,[(r(!0),d(y,null,P(j.value,(e,i)=>(r(),d(y,{key:i},[e.length!==0?(r(),d("li",ie,[s(h,{to:{name:"policies-list-view",params:{policyPath:e.path}}},{default:o(()=>[l(g(e.name)+": "+g(e.length),1)]),_:2},1032,["to"])])):S("",!0)],64))),128))])]),_:1},8,["term"])]),_:1})])]),_:1})):S("",!0),l(),s(m(x),null,{body:o(()=>{var e;return[s(ee,{id:"code-block-mesh","resource-fetcher":z,"resource-fetcher-watch-key":((e=n.value)==null?void 0:e.name)||null},null,8,["resource-fetcher-watch-key"])]}),_:1}),l(),s(Y)])]),_:1})]),_:1})}}});const De=K(ue,[["__scopeId","data-v-0f52eab1"]]);export{De as default};
