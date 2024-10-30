import{_ as F}from"./PolicySummary.vue_vue_type_script_setup_true_lang-DfI3Ne7G.js";import{_ as P}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CK_GdeMA.js";import{d as T,e as t,o as n,m as p,w as o,a as s,k as d,t as _,b as u,c as v,H as D,J as $,p as h,q as b}from"./index-CgC5RQPZ.js";const A=T({__name:"PolicySummaryView",props:{items:{},policyType:{}},setup(f){const l=f;return(B,M)=>{const g=t("XEmptyState"),C=t("RouteTitle"),x=t("XAction"),w=t("DataSource"),E=t("AppView"),S=t("DataCollection"),V=t("RouteView");return n(),p(V,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:i})=>[s(S,{items:l.items,predicate:m=>m.id===e.params.policy,find:!0},{empty:o(()=>[s(g,null,{title:o(()=>[d("h2",null,_(i("common.collection.summary.empty_title",{type:l.policyType.name})),1)]),default:o(()=>[u(),d("p",null,_(i("common.collection.summary.empty_message",{type:l.policyType.name})),1)]),_:2},1024)]),default:o(({items:m})=>[(n(!0),v(D,null,$([m[0]],c=>(n(),p(E,{key:c.id},{title:o(()=>[d("h2",null,[s(x,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:o(()=>[s(C,{title:i("policies.routes.item.title",{name:c.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:o(()=>[u(),c?(n(),p(F,{key:0,policy:c},{default:o(()=>[s(P,{resource:c.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:R})=>[R?(n(),p(w,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:r=>{a(y=>y(r))},onError:r=>{a((y,k)=>k(r))}},null,8,["src","onChange","onError"])):h("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["policy"])):h("",!0)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),Q=b(A,[["__scopeId","data-v-a4c6b92f"]]);export{Q as default};
