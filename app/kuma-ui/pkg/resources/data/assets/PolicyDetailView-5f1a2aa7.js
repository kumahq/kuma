import{d as v,k as b,A as F,a as i,o as a,c as m,e as t,l as B,f as d,w as s,b as p,F as N,C as S,t as w,aq as L,m as $,p as T}from"./index-933c8957.js";import{_ as q}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-57e73b2f.js";import{E as V}from"./ErrorBlock-840882fa.js";import{_ as P}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-21f5d8d7.js";import{_ as I}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-12338d91.js";import{T as K}from"./TextWithCopyButton-6da05642.js";import"./index-fce48c05.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-d0d3d9cc.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-12f9cd74.js";import"./uniqueId-90cc9b93.js";import"./CopyButton-4ca50dda.js";import"./toYaml-4e00099e.js";const M={key:3,"data-testid":"affected-data-plane-proxies"},A=v({__name:"PolicyConnections",props:{mesh:{},policyPath:{},policyName:{}},setup(g){const{t:x}=b(),h=g,r=F("");return(C,c)=>{const u=i("KInput"),y=i("RouterLink");return a(),m("div",null,[t(u,{id:"dataplane-search",modelValue:r.value,"onUpdate:modelValue":c[0]||(c[0]=e=>r.value=e),type:"text",placeholder:B(x)("policies.detail.dataplane_input_placeholder"),required:"","data-testid":"dataplane-search-input"},null,8,["modelValue","placeholder"]),d(),t(L,{src:`/meshes/${h.mesh}/policy-path/${h.policyPath}/policy/${h.policyName}/dataplanes`},{default:s(({data:e,error:n})=>[n?(a(),p(V,{key:0,error:n},null,8,["error"])):e===void 0?(a(),p(P,{key:1})):e.items.length===0?(a(),p(q,{key:2})):(a(),m("ul",M,[(a(!0),m(N,null,S(e.items.filter(o=>o.name.toLowerCase().includes(r.value.toLowerCase())),(o,_)=>(a(),m("li",{key:_,"data-testid":"dataplane-name"},[t(y,{to:{name:"data-plane-detail-view",params:{mesh:o.mesh,dataPlane:o.name}}},{default:s(()=>[d(w(o.name),1)]),_:2},1032,["to"])]))),128))]))]),_:1},8,["src"])])}}}),D={key:2,class:"stack","data-testid":"detail-view-details"},ee=v({__name:"PolicyDetailView",setup(g){return(x,h)=>{const r=i("RouteTitle"),C=i("KCard"),c=i("DataSource"),u=i("AppView"),y=i("RouteView");return a(),p(y,{name:"policy-detail-view",params:{mesh:"",policy:"",policyPath:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:s(({route:e,t:n})=>[t(u,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"policy-list-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath}},text:n("policies.routes.item.breadcrumbs")}]},{title:s(()=>[$("h1",null,[t(K,{text:e.params.policy},{default:s(()=>[t(r,{title:n("policies.routes.item.title",{name:e.params.policy})},null,8,["title"])]),_:2},1032,["text"])])]),default:s(()=>[d(),t(c,{src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}`},{default:s(({data:o,error:_})=>[_?(a(),p(V,{key:0,error:_},null,8,["error"])):o===void 0?(a(),p(P,{key:1})):(a(),m("div",D,[t(C,null,{default:s(()=>[$("h2",null,w(n("policies.detail.affected_dpps")),1),d(),t(A,{class:"mt-4",mesh:e.params.mesh,"policy-name":e.params.policy,"policy-path":e.params.policyPath},null,8,["mesh","policy-name","policy-path"])]),_:2},1024),d(),t(I,{resource:o.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:l=>e.update({codeSearch:l}),onFilterModeChange:l=>e.update({codeFilter:l}),onRegExpModeChange:l=>e.update({codeRegExp:l})},{default:s(({copy:l,copying:R})=>[R?(a(),p(c,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:f=>{l(k=>k(f))},onError:f=>{l((k,E)=>E(f))}},null,8,["src","onChange","onError"])):T("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{ee as default};
