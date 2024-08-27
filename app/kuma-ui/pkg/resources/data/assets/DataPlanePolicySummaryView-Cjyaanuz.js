import{_ as R}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BkeCm4yq.js";import{_ as w}from"./PolicySummary.vue_vue_type_script_setup_true_lang-Be5NwiTy.js";import{d as V,r as s,o as p,m as n,w as o,b as t,k as P,e as k,p as m,q as $}from"./index-Is4zmHdk.js";import"./CodeBlock-DvLuvw_5.js";const E=V({__name:"DataPlanePolicySummaryView",setup(S){return(D,v)=>{const d=s("RouteTitle"),_=s("RouterLink"),i=s("DataSource"),h=s("DataLoader"),y=s("AppView"),u=s("RouteView");return p(),n(u,{name:"data-plane-policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:f})=>[t(i,{src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}`},{default:o(({data:r,error:g})=>[t(y,null,{title:o(()=>[P("h2",null,[t(_,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:o(()=>[t(d,{title:f("policies.routes.item.title",{name:e.params.policy})},null,8,["title"])]),_:2},1032,["to"])])]),default:o(()=>[k(),t(h,{data:[r],errors:[g]},{default:o(()=>[r?(p(),n(w,{key:0,policy:r},{default:o(()=>[t(R,{resource:r.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:x})=>[x?(p(),n(i,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:c=>{a(l=>l(c))},onError:c=>{a((l,C)=>C(c))}},null,8,["src","onChange","onError"])):m("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["policy"])):m("",!0)]),_:2},1032,["data","errors"])]),_:2},1024)]),_:2},1032,["src"])]),_:1})}}}),N=$(E,[["__scopeId","data-v-98814b05"]]);export{N as default};
