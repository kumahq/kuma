import{d as S,l as V,u as C,a as m,o as n,b as y,w as a,e as _,p as o,f as t,q as s,t as c,c as u,a6 as B,v as h,F as x,B as P,C as T,a1 as b,_ as I}from"./index-646486ee.js";import{_ as A}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-f5640fed.js";import{_ as N}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-1b81a712.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-a2e2bc37.js";import"./toYaml-4e00099e.js";const q=l=>(P("data-v-e1e1efa1"),l=l(),T(),l),K={class:"summary-title-wrapper"},D=q(()=>o("img",{"aria-hidden":"true",src:b},null,-1)),E={class:"summary-title"},F={key:1,class:"stack"},L={key:0},Q={class:"mt-4 stack"},$={key:0},j={class:"mt-4"},z=S({__name:"PolicySummaryView",props:{name:{},policy:{default:void 0},policyType:{}},setup(l){const{t:i}=V(),f=C(),e=l;return(G,H)=>{const g=m("RouteTitle"),v=m("RouterLink"),k=m("KBadge"),w=m("AppView"),R=m("RouteView");return n(),y(R,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:""}},{default:a(({route:r})=>[_(w,null,{title:a(()=>[o("div",K,[D,t(),o("h2",E,[_(v,{to:{name:"policy-detail-view",params:{policy:e.name}}},{default:a(()=>[_(g,{title:s(i)("policies.routes.item.title",{name:e.name})},null,8,["title"])]),_:1},8,["to"])])])]),default:a(()=>{var d;return[t(),e.policy===void 0?(n(),y(A,{key:0},{message:a(()=>[o("p",null,c(s(i)("common.collection.summary.empty_message",{type:e.policyType.name})),1)]),default:a(()=>[t(c(s(i)("common.collection.summary.empty_title",{type:e.policyType.name}))+" ",1)]),_:1})):(n(),u("div",F,[(d=e.policy.spec)!=null&&d.targetRef?(n(),u("div",L,[o("h3",null,c(s(i)("policies.routes.item.overview")),1),t(),o("div",Q,[_(B,null,{title:a(()=>[t(c(s(i)("http.api.property.targetRef")),1)]),body:a(()=>{var p;return[(p=e.policy.spec)!=null&&p.targetRef?(n(),y(k,{key:0,appearance:"neutral"},{default:a(()=>[t(c(e.policy.spec.targetRef.kind),1),e.policy.spec.targetRef.name?(n(),u("span",$,[t(":"),o("b",null,c(e.policy.spec.targetRef.name),1)])):h("",!0)]),_:1})):(n(),u(x,{key:1},[t(c(s(i)("common.detail.none")),1)],64))]}),_:1})])])):h("",!0),t(),o("div",null,[o("h3",null,c(s(i)("policies.routes.item.config")),1),t(),o("div",j,[_(N,{id:"code-block-policy",resource:e.policy,"resource-fetcher":p=>s(f).getSinglePolicyEntity({name:r.params.policy,mesh:r.params.mesh,path:r.params.policyPath},p),"is-searchable":"",query:r.params.codeSearch,onQueryChange:p=>r.update({codeSearch:p})},null,8,["resource","resource-fetcher","query","onQueryChange"])])])]))]}),_:2},1024)]),_:1})}}});const X=I(z,[["__scopeId","data-v-e1e1efa1"]]);export{X as default};
