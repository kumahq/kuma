import{_ as b}from"./PolicySummary.vue_vue_type_script_setup_true_lang-WlLW3Vtb.js";import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CgraeAfw.js";import{d as P,e as t,o as n,m as p,w as o,a as s,k as d,t as _,b as u,c as T,H as v,J as D,p as h,q as $}from"./index-Bo5vSFZC.js";import"./CodeBlock-BVhuix4S.js";const B=P({__name:"PolicySummaryView",props:{items:{},policyType:{}},setup(f){const i=f;return(M,N)=>{const g=t("XEmptyState"),C=t("RouteTitle"),x=t("RouterLink"),R=t("DataSource"),k=t("AppView"),w=t("DataCollection"),E=t("RouteView");return n(),p(E,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:l})=>[s(w,{items:i.items,predicate:r=>r.id===e.params.policy,find:!0},{empty:o(()=>[s(g,null,{title:o(()=>[d("h2",null,_(l("common.collection.summary.empty_title",{type:i.policyType.name})),1)]),default:o(()=>[u(),d("p",null,_(l("common.collection.summary.empty_message",{type:i.policyType.name})),1)]),_:2},1024)]),default:o(({items:r})=>[(n(!0),T(v,null,D([r[0]],c=>(n(),p(k,{key:c.id},{title:o(()=>[d("h2",null,[s(x,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:o(()=>[s(C,{title:l("policies.routes.item.title",{name:c.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:o(()=>[u(),c?(n(),p(b,{key:0,policy:c},{default:o(()=>[s(F,{resource:c.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:S})=>[S?(n(),p(R,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:m=>{a(y=>y(m))},onError:m=>{a((y,V)=>V(m))}},null,8,["src","onChange","onError"])):h("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["policy"])):h("",!0)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),X=$(B,[["__scopeId","data-v-6b3be19b"]]);export{X as default};
