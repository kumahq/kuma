import{d as F,f as T,g as d,h as Y,o as l,c as r,a as h,u as a,b as o,k as X,cv as K,bP as M,cw as U,cm as O,c8 as R,cc as W,w as b,e as u,F as k,cd as S,bV as m,i as $,c3 as N,j as B,bX as q,bY as z}from"./index.bd548025.js";import{C as G}from"./ContentWrapper.92bb9d03.js";import{_ as D,a as A,M as H}from"./MeshResources.7212245a.js";import{_ as J}from"./LabelList.vue_vue_type_style_index_0_lang.911f7b12.js";import{_ as Q}from"./YamlView.vue_vue_type_script_setup_true_lang.bc9d3e4b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.0d00632b.js";import"./ErrorBlock.ee0cc1df.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.f0102383.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.867fc141.js";import"./_commonjsHelpers.f037b798.js";const Z={class:"chart-container mt-24"},ee=F({__name:"MeshCharts",setup(v){const t=T(),f=d(()=>t.getters.getServiceResourcesFetching),p=d(()=>t.getters.getMeshInsightsFetching),y=d(()=>t.getters.getChart("services")),g=d(()=>t.getters.getChart("dataplanes")),c=d(()=>t.getters.getChart("policies")),P=d(()=>t.getters.getChart("kumaDPVersions")),w=d(()=>t.getters.getChart("envoyVersions"));Y(()=>t.state.selectedMesh,function(){I()}),I();function I(){t.dispatch("fetchMeshInsights",t.state.selectedMesh),t.dispatch("fetchServices",t.state.selectedMesh)}return(E,x)=>(l(),r("div",Z,[h(D,{class:"chart",title:{singular:"SERVICE",plural:"SERVICES"},data:a(y).data,"is-loading":a(f),"save-chart":""},null,8,["data","is-loading"]),o(),h(D,{class:"chart",title:{singular:"DP PROXY",plural:"DP PROXIES"},data:a(g).data,url:{name:"data-plane-list-view",params:{mesh:a(t).state.selectedMesh}},"is-loading":a(p)},null,8,["data","url","is-loading"]),o(),h(D,{class:"chart",title:{singular:"POLICY",plural:"POLICIES"},data:a(c).data,url:{name:"policies",params:{mesh:a(t).state.selectedMesh}},"is-loading":a(p)},null,8,["data","url","is-loading"]),o(),h(A,{class:"chart",title:"KUMA DP",data:a(P).data,"is-loading":a(p)},null,8,["data","is-loading"]),o(),h(A,{class:"chart",title:"ENVOY",data:a(w).data,"is-loading":a(p),"display-am-charts-logo":""},null,8,["data","is-loading"])]))}});const ae=X(ee,[["__scopeId","data-v-2a02347c"]]),te=v=>(q("data-v-2d55bce4"),v=v(),z(),v),se={key:0},ne={key:1},le={key:1},oe={class:"policy-counts"},ce=te(()=>u("h4",null,`
                  Policies
                `,-1)),re={key:0},ie={key:0,class:"mt-4"},ue=F({__name:"MeshOverviewView",setup(v){const t=K(),f=T(),p=M(!0),y=M(!1),g=M(!1),c=M(null),P=M(null),w=d(()=>c.value!==null?U(c.value):null),I=d(()=>{if(c.value===null)return null;const{name:s,type:n,creationTime:_,modificationTime:e}=c.value;return{name:s,type:n,created:O(_),modified:O(e),"Data Plane Proxies":f.state.meshInsight.dataplanes.total}}),E=d(()=>{var C;if(c.value===null)return null;const s=V(c.value,"mtls"),n=V(c.value,"logging"),_=V(c.value,"metrics"),e=V(c.value,"tracing"),i=Boolean((C=c.value.routing)==null?void 0:C.localityAwareLoadBalancing);return{mtls:s,logging:n,metrics:_,tracing:e,localityAwareLoadBalancing:i}}),x=d(()=>f.state.policyTypes.map(s=>{var n,_;return{...s,length:(_=(n=f.state.meshInsight.policies[s.name])==null?void 0:n.total)!=null?_:0}}));Y(()=>t.params.mesh,function(){t.name==="single-mesh-overview"&&(p.value=!0,g.value=!1,y.value=!1,L())}),L();async function L(){p.value=!0,g.value=!1;const s=t.params.mesh;try{c.value=await R.getMesh({name:s}),P.value=await R.getMeshInsights({name:s})}catch(n){y.value=!0,g.value=!0,console.error(n)}finally{p.value=!1}}function V(s,n){var i,C;if(s===null||s[n]===void 0)return!1;const _=(C=(i=s[n].enabledBackend)!=null?i:s[n].defaultBackend)!=null?C:s[n].backends[0].name,e=s[n].backends.find(j=>j.name===_);return`${e.type} / ${e.name}`}return(s,n)=>{const _=W("router-link");return l(),r(k,null,[h(ae),o(),h(G,{class:"mt-8"},{content:b(()=>[c.value!==null?(l(),r("div",se,[h(J,{"has-error":y.value,"is-loading":p.value,"is-empty":g.value},{default:b(()=>[u("div",null,[u("ul",null,[(l(!0),r(k,null,S(a(I),(e,i)=>(l(),r("li",{key:i},[u("h4",null,m(i),1),o(),typeof e=="boolean"?(l(),$(a(N),{key:0,appearance:e?"success":"danger"},{default:b(()=>[o(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),r("p",ne,m(e),1))]))),128))])]),o(),u("div",null,[u("ul",null,[(l(!0),r(k,null,S(a(E),(e,i)=>(l(),r("li",{key:i},[u("h4",null,m(i),1),o(),typeof e=="boolean"?(l(),$(a(N),{key:0,appearance:e?"success":"danger"},{default:b(()=>[o(m(e?"Enabled":"Disabled"),1)]),_:2},1032,["appearance"])):(l(),r("p",le,m(e),1))]))),128))])]),o(),u("div",null,[u("ul",oe,[u("li",null,[ce,o(),u("ul",null,[(l(!0),r(k,null,S(a(x),(e,i)=>(l(),r(k,{key:i},[e.length!==0?(l(),r("li",re,[h(_,{to:{name:e.path}},{default:b(()=>[o(m(e.name)+": "+m(e.length),1)]),_:2},1032,["to"])])):B("",!0)],64))),128))])])])])]),_:1},8,["has-error","is-loading","is-empty"])])):B("",!0)]),_:1}),o(),a(w)!==null?(l(),r("div",ie,[h(Q,{id:"code-block-mesh",content:a(w)},null,8,["content"])])):B("",!0),o(),h(H,{class:"mt-6"})],64)}}});const be=X(ue,[["__scopeId","data-v-2d55bce4"]]);export{be as default};
