import{d as S,g as V,Q as C,r as m,o as n,i as y,w as a,j as _,p as o,n as t,k as s,H as c,l as u,a7 as x,m as h,F as B,D as P,G as T,a2 as I,t as b}from"./index-a6f5023f.js";import{_ as A}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-1974ccfb.js";import{_ as N}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-3ee102e1.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ad731d3d.js";import"./toYaml-4e00099e.js";const D=l=>(P("data-v-5391a103"),l=l(),T(),l),K={class:"summary-title-wrapper"},Q=D(()=>o("img",{"aria-hidden":"true",src:I},null,-1)),q={class:"summary-title"},E={key:1,class:"stack"},F={key:0},L={class:"mt-4 stack"},$={key:0},j={class:"mt-4"},G=S({__name:"PolicySummaryView",props:{name:{},policy:{default:void 0},policyType:{}},setup(l){const{t:i}=V(),f=C(),e=l;return(H,z)=>{const g=m("RouteTitle"),k=m("RouterLink"),v=m("KBadge"),w=m("AppView"),R=m("RouteView");return n(),y(R,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:""}},{default:a(({route:r})=>[_(w,null,{title:a(()=>[o("div",K,[Q,t(),o("h2",q,[_(k,{to:{name:"policy-detail-view",params:{policy:e.name}}},{default:a(()=>[_(g,{title:s(i)("policies.routes.item.title",{name:e.name}),render:!0},null,8,["title"])]),_:1},8,["to"])])])]),default:a(()=>{var d;return[t(),e.policy===void 0?(n(),y(A,{key:0},{message:a(()=>[o("p",null,c(s(i)("common.collection.summary.empty_message",{type:e.policyType.name})),1)]),default:a(()=>[t(c(s(i)("common.collection.summary.empty_title",{type:e.policyType.name}))+" ",1)]),_:1})):(n(),u("div",E,[(d=e.policy.spec)!=null&&d.targetRef?(n(),u("div",F,[o("h3",null,c(s(i)("policies.routes.item.overview")),1),t(),o("div",L,[_(x,null,{title:a(()=>[t(c(s(i)("http.api.property.targetRef")),1)]),body:a(()=>{var p;return[(p=e.policy.spec)!=null&&p.targetRef?(n(),y(v,{key:0,appearance:"neutral"},{default:a(()=>[t(c(e.policy.spec.targetRef.kind),1),e.policy.spec.targetRef.name?(n(),u("span",$,[t(":"),o("b",null,c(e.policy.spec.targetRef.name),1)])):h("",!0)]),_:1})):(n(),u(B,{key:1},[t(c(s(i)("common.detail.none")),1)],64))]}),_:1})])])):h("",!0),t(),o("div",null,[o("h3",null,c(s(i)("policies.routes.item.config")),1),t(),o("div",j,[_(N,{id:"code-block-policy",resource:e.policy,"resource-fetcher":p=>s(f).getSinglePolicyEntity({name:r.params.policy,mesh:r.params.mesh,path:r.params.policyPath},p),"is-searchable":"",query:r.params.codeSearch,onQueryChange:p=>r.update({codeSearch:p})},null,8,["resource","resource-fetcher","query","onQueryChange"])])])]))]}),_:2},1024)]),_:1})}}});const X=b(G,[["__scopeId","data-v-5391a103"]]);export{X as default};
