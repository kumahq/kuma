import{_ as w}from"./PolicySummary.vue_vue_type_script_setup_true_lang-QQ18Hg6S.js";import{_ as V}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-7LOb7PdN.js";import{d as P,e as s,o as r,m as n,w as o,a as t,k as R,b as $,p as m,q as E}from"./index-Yqc5mH7h.js";const S=P({__name:"DataPlanePolicySummaryView",setup(k){return(D,v)=>{const d=s("RouteTitle"),_=s("XAction"),i=s("DataSource"),h=s("DataLoader"),y=s("AppView"),f=s("RouteView");return r(),n(f,{name:"data-plane-policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:u})=>[t(i,{src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}`},{default:o(({data:c,error:g})=>[t(y,null,{title:o(()=>[R("h2",null,[t(_,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:o(()=>[t(d,{title:u("policies.routes.item.title",{name:e.params.policy})},null,8,["title"])]),_:2},1032,["to"])])]),default:o(()=>[$(),t(h,{data:[c],errors:[g]},{default:o(()=>[c?(r(),n(w,{key:0,policy:c},{default:o(()=>[t(V,{resource:c.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:x})=>[x?(r(),n(i,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:p=>{a(l=>l(p))},onError:p=>{a((l,C)=>C(p))}},null,8,["src","onChange","onError"])):m("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["policy"])):m("",!0)]),_:2},1032,["data","errors"])]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}}),N=E(S,[["__scopeId","data-v-afddaea7"]]);export{N as default};
