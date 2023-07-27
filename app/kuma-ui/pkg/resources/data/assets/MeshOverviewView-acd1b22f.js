import{d as q,u as F,j as f,c as v,o as u,e as y,h as s,g as c,r as G,a as h,w as i,b as p,q as T,J as O,F as w,s as S,L as E,t as k,f as H}from"./index-065c0e80.js";import{g as J,n as W,s as K,q as Q,D as M,f as z,e as U,r as R,A as X,_ as Y}from"./RouteView.vue_vue_type_script_setup_true_lang-1d679e8a.js";import{_ as Z}from"./RouteTitle.vue_vue_type_script_setup_true_lang-44fde05c.js";import{D as N,a as j}from"./DefinitionListItem-adce3dd6.js";import{_ as ee}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-a9d785f8.js";import{_ as te}from"./StatusInfo.vue_vue_type_script_setup_true_lang-d1fde56e.js";import{T as ae}from"./TextWithCopyButton-3f72ac03.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-eb21505c.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-cc00a44f.js";import"./ErrorBlock-9add341a.js";const ne={class:"chart-box-list"},le=q({__name:"MeshCharts",setup(D){const o=J(),b=W(),x=F(),C=f(!1),_=f({total:0,online:0,partiallyDegraded:0,offline:0}),m=f({total:0,internal:0,external:0}),l=f({kumaDp:{},envoy:{}}),d=v(()=>{const e=[],{internal:t,external:a}=m.value;return t&&e.push({title:o.t("common.charts.services.internalLabel"),data:t}),a&&e.push({title:o.t("common.charts.services.externalLabel"),data:a}),{title:o.t("common.charts.services.title"),showTotal:!0,dataPoints:e}}),V=v(()=>{const e=[],{total:t,online:a,partiallyDegraded:n}=_.value;if(t>0){e.push({title:o.t("http.api.value.online"),statusKeyword:"online",data:a}),n>0&&e.push({title:o.t("http.api.value.partially_degraded"),statusKeyword:"partially_degraded",data:n});const r=t-n-a;r>0&&e.push({title:o.t("http.api.value.offline"),statusKeyword:"offline",data:r})}return{title:o.t("common.charts.dataPlaneProxies.title"),showTotal:!0,dataPoints:e}}),B=v(()=>{const e=Object.entries(l.value.kumaDp).map(([t,a])=>({title:t,data:a.total??0}));return e.sort((t,a)=>t.title==="unknown"?1:a.title==="unknown"?-1:K(t.title,a.title)),{title:o.t("common.charts.kumaDp.title"),subtitle:o.t("common.charts.kumaDp.subtitle"),dataPoints:e}}),P=v(()=>{const e=Object.entries(l.value.envoy).map(([t,a])=>({title:t,data:a.total??0}));return e.sort((t,a)=>t.title==="unknown"?1:a.title==="unknown"?-1:K(t.title,a.title)),{title:o.t("common.charts.envoy.title"),subtitle:o.t("common.charts.envoy.subtitle"),dataPoints:e}});L();async function L(){C.value=!0;const e=x.params.mesh;try{const t=await b.getMeshInsights({name:e}),a=Q([t]);$(a),g(a),I(a)}catch{_.value={total:0,online:0,partiallyDegraded:0,offline:0},m.value={total:0,internal:0,external:0},l.value={kumaDp:{},envoy:{}}}finally{C.value=!1}}function $(e){const{total:t,online:a,partiallyDegraded:n}=e.dataplanes;_.value={total:t,online:a,partiallyDegraded:n,offline:t-a-n}}function g(e){const{total:t,internal:a,external:n}=e.services;m.value={total:t,internal:a,external:n}}function I(e){l.value=e.dpVersions}return(e,t)=>(u(),y("div",ne,[s(M,{data:d.value},null,8,["data"]),c(),s(M,{data:V.value},null,8,["data"]),c(),s(M,{data:B.value},null,8,["data"]),c(),s(M,{data:P.value},null,8,["data"])]))}});const se=z(le,[["__scopeId","data-v-375c50a1"]]);function oe(D){return D!=null}const re={class:"stack"},ie={class:"columns"},ue=q({__name:"MeshOverviewView",setup(D){const{t:o}=J(),b=W(),x=F(),C=U(),_=f(!0),m=f(null),l=f(null),d=f(null),V=v(()=>{if(l.value===null||d.value===null)return null;const{name:e,creationTime:t,modificationTime:a}=l.value;return{name:e,created:R(t),modified:R(a),"Data Plane Proxies":d.value.dataplanes.total}}),B=v(()=>{var A;if(l.value===null)return null;const e=g(l.value,"mtls"),t=g(l.value,"logging"),a=g(l.value,"metrics"),n=g(l.value,"tracing"),r=!!((A=l.value.routing)!=null&&A.localityAwareLoadBalancing);return{mtls:e,logging:t,metrics:a,tracing:n,localityAwareLoadBalancing:r}}),P=v(()=>d.value===null?0:Object.values(d.value.policies).reduce((e,t)=>e+t.total,0)),L=v(()=>d.value===null?[]:Object.entries(d.value.policies).map(([e,t])=>{const a=C.state.policyTypesByName[e];return a&&t.total!==0?{name:a.name,path:a.path,total:t.total}:null}).filter(oe));$();async function $(){_.value=!0,m.value=null;const e=x.params.mesh;try{l.value=await b.getMesh({name:e}),d.value=await b.getMeshInsights({name:e})}catch(t){t instanceof Error?m.value=t:console.error(m),l.value=null,d.value=null}finally{_.value=!1}}function g(e,t){if(e===null||e[t]===void 0)return!1;const a=e[t].enabledBackend??e[t].defaultBackend??e[t].backends[0].name,n=e[t].backends.find(r=>r.name===a);return`${n.type} / ${n.name}`}async function I(e){const t=x.params.mesh;return await b.getMesh({name:t},e)}return(e,t)=>{const a=G("router-link");return u(),h(Y,null,{default:i(()=>[s(Z,{title:p(o)("meshes.routes.overview.title")},null,8,["title"]),c(),s(X,null,{default:i(()=>[T("div",re,[s(p(O),null,{body:i(()=>[s(se)]),_:1}),c(),l.value!==null?(u(),h(p(O),{key:0},{body:i(()=>[T("div",ie,[s(te,{"is-loading":_.value,error:m.value,"is-empty":l.value===null||d.value===null},{default:i(()=>[s(N,null,{default:i(()=>[(u(!0),y(w,null,S(V.value,(n,r)=>(u(),h(j,{key:r,term:p(o)(`http.api.property.${r}`)},{default:i(()=>[typeof n=="boolean"?(u(),h(p(E),{key:0,appearance:n?"success":"danger"},{default:i(()=>[c(k(n?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):r==="name"&&typeof n=="string"?(u(),h(ae,{key:1,text:n},null,8,["text"])):(u(),y(w,{key:2},[c(k(n),1)],64))]),_:2},1032,["term"]))),128))]),_:1})]),_:1},8,["is-loading","error","is-empty"]),c(),s(N,null,{default:i(()=>[(u(!0),y(w,null,S(B.value,(n,r)=>(u(),h(j,{key:r,term:p(o)(`http.api.property.${r}`)},{default:i(()=>[typeof n=="boolean"?(u(),h(p(E),{key:0,appearance:n?"success":"danger"},{default:i(()=>[c(k(n?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(u(),y(w,{key:1},[c(k(n),1)],64))]),_:2},1032,["term"]))),128))]),_:1}),c(),s(N,null,{default:i(()=>[s(j,{term:`Policies (${P.value})`},{default:i(()=>[T("ul",null,[(u(!0),y(w,null,S(L.value,(n,r)=>(u(),y("li",{key:r},[s(a,{to:{name:"policies-list-view",params:{policyPath:n.path}}},{default:i(()=>[c(k(n.name)+": "+k(n.total),1)]),_:2},1032,["to"])]))),128))])]),_:1},8,["term"])]),_:1})])]),_:1})):H("",!0),c(),s(p(O),null,{body:i(()=>{var n;return[s(ee,{id:"code-block-mesh","resource-fetcher":I,"resource-fetcher-watch-key":((n=l.value)==null?void 0:n.name)||null},null,8,["resource-fetcher-watch-key"])]}),_:1})])]),_:1})]),_:1})}}});const ke=z(ue,[["__scopeId","data-v-252e58a8"]]);export{ke as default};
