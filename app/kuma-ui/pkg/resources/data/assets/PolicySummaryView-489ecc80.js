import{d as x,l as C,u as S,a as m,o as l,b as u,w as i,e as d,p as a,f as o,q as c,t as p,c as _,a0 as V,s as h,F as E,z as F,A as B,X as P,_ as T}from"./index-f56c27ab.js";import{_ as b}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-2656c87a.js";import{_ as A}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-338e6d95.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-68115bc8.js";import"./CopyButton-7b31a54c.js";import"./index-52545d1d.js";import"./toYaml-4e00099e.js";const I=r=>(F("data-v-8393e055"),r=r(),B(),r),M={class:"summary-title-wrapper"},N=I(()=>a("img",{"aria-hidden":"true",src:P},null,-1)),q={class:"summary-title"},K={key:1,class:"stack"},D={key:0},L={class:"mt-4 stack"},Q={key:0},$={class:"mt-4"},z=x({__name:"PolicySummaryView",props:{name:{},policy:{default:void 0},policyType:{}},setup(r){const{t:n}=C(),f=S(),e=r;return(X,j)=>{const g=m("RouteTitle"),v=m("RouterLink"),R=m("KBadge"),k=m("AppView"),w=m("RouteView");return l(),u(w,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:i(({route:s})=>[d(k,null,{title:i(()=>[a("div",M,[N,o(),a("h2",q,[d(v,{to:{name:"policy-detail-view",params:{policy:e.name}}},{default:i(()=>[d(g,{title:c(n)("policies.routes.item.title",{name:e.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:i(()=>{var y;return[o(),e.policy===void 0?(l(),u(b,{key:0},{message:i(()=>[a("p",null,p(c(n)("common.collection.summary.empty_message",{type:e.policyType.name})),1)]),default:i(()=>[o(p(c(n)("common.collection.summary.empty_title",{type:e.policyType.name}))+" ",1)]),_:1})):(l(),_("div",K,[(y=e.policy.spec)!=null&&y.targetRef?(l(),_("div",D,[a("h3",null,p(c(n)("policies.routes.item.overview")),1),o(),a("div",L,[d(V,null,{title:i(()=>[o(p(c(n)("http.api.property.targetRef")),1)]),body:i(()=>{var t;return[(t=e.policy.spec)!=null&&t.targetRef?(l(),u(R,{key:0,appearance:"neutral"},{default:i(()=>[o(p(e.policy.spec.targetRef.kind),1),e.policy.spec.targetRef.name?(l(),_("span",Q,[o(":"),a("b",null,p(e.policy.spec.targetRef.name),1)])):h("",!0)]),_:1})):(l(),_(E,{key:1},[o(p(c(n)("common.detail.none")),1)],64))]}),_:1})])])):h("",!0),o(),a("div",null,[a("h3",null,p(c(n)("policies.routes.item.config")),1),o(),a("div",$,[d(A,{id:"code-block-policy",resource:e.policy,"resource-fetcher":t=>c(f).getSinglePolicyEntity({name:s.params.policy,mesh:s.params.mesh,path:s.params.policyPath},t),"is-searchable":"",query:s.params.codeSearch,"is-filter-mode":s.params.codeFilter==="true","is-reg-exp-mode":s.params.codeRegExp==="true",onQueryChange:t=>s.update({codeSearch:t}),onFilterModeChange:t=>s.update({codeFilter:t}),onRegExpModeChange:t=>s.update({codeRegExp:t})},null,8,["resource","resource-fetcher","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]))]}),_:2},1024)]),_:1})}}});const Z=T(z,[["__scopeId","data-v-8393e055"]]);export{Z as default};
