import{_ as S}from"./PolicySummary.vue_vue_type_script_setup_true_lang-Qdj2FMoW.js";import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CsUbsfgu.js";import{d as P,a as t,o as c,b as i,w as a,e as n,a9 as T,f as d,t as y,m as u,c as $,F as v,G as D,q as h,_ as B}from"./index-DzXsKW1h.js";import"./CodeBlock-CqYy31tT.js";import"./toYaml-DB9FPXFY.js";const M=P({__name:"PolicySummaryView",props:{items:{},policyType:{}},setup(f){const p=f;return(N,b)=>{const g=t("RouteTitle"),C=t("RouterLink"),x=t("DataSource"),R=t("AppView"),w=t("DataCollection"),V=t("RouteView");return c(),i(V,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:r})=>[n(w,{items:p.items,predicate:l=>l.id===e.params.policy,find:!0},{empty:a(()=>[n(T,null,{title:a(()=>[d(y(r("common.collection.summary.empty_title",{type:p.policyType.name})),1)]),default:a(()=>[d(),u("p",null,y(r("common.collection.summary.empty_message",{type:p.policyType.name})),1)]),_:2},1024)]),default:a(({items:l})=>[(c(!0),$(v,null,D([l[0]],s=>(c(),i(R,{key:s.id},{title:a(()=>[u("h2",null,[n(C,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:a(()=>[n(g,{title:r("policies.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:a(()=>[d(),s?(c(),i(S,{key:0,policy:s},{default:a(()=>[n(F,{resource:s.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:a(({copy:o,copying:k})=>[k?(c(),i(x,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:m=>{o(_=>_(m))},onError:m=>{o((_,E)=>E(m))}},null,8,["src","onChange","onError"])):h("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["policy"])):h("",!0)]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),I=B(M,[["__scopeId","data-v-3c2a468a"]]);export{I as default};
