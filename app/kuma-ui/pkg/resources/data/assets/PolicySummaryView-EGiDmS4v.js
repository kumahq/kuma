import{_ as E}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-L1J1bxeo.js";import{d as S,a as l,o as c,b as m,w as a,e as r,m as i,f as s,a2 as b,t as n,Z as F,c as d,p as _,F as B,_ as P}from"./index-pAyRVwwQ.js";import"./CodeBlock-6c7dCnil.js";import"./toYaml-sPaYOD3i.js";const T={key:1,class:"stack"},$={key:0},D={class:"mt-4 stack"},M={key:0},N={class:"mt-4"},q=S({__name:"PolicySummaryView",props:{policy:{default:void 0},policyType:{}},setup(f){const t=f;return(A,K)=>{const g=l("RouteTitle"),R=l("RouterLink"),k=l("KBadge"),v=l("DataSource"),C=l("AppView"),x=l("RouteView");return c(),m(x,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:p})=>[r(C,null,{title:a(()=>[i("h2",null,[r(R,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:a(()=>[r(g,{title:p("policies.routes.item.title",{name:e.params.policy})},null,8,["title"])]),_:2},1032,["to"])])]),default:a(()=>{var u;return[s(),t.policy===void 0?(c(),m(b,{key:0},{message:a(()=>[i("p",null,n(p("common.collection.summary.empty_message",{type:t.policyType.name})),1)]),default:a(()=>[s(n(p("common.collection.summary.empty_title",{type:t.policyType.name}))+" ",1)]),_:2},1024)):(c(),d("div",T,[(u=t.policy.spec)!=null&&u.targetRef?(c(),d("div",$,[i("h3",null,n(p("policies.routes.item.overview")),1),s(),i("div",D,[r(F,null,{title:a(()=>[s(n(p("http.api.property.targetRef")),1)]),body:a(()=>{var o;return[(o=t.policy.spec)!=null&&o.targetRef?(c(),m(k,{key:0,appearance:"neutral"},{default:a(()=>[s(n(t.policy.spec.targetRef.kind),1),t.policy.spec.targetRef.name?(c(),d("span",M,[s(":"),i("b",null,n(t.policy.spec.targetRef.name),1)])):_("",!0)]),_:1})):(c(),d(B,{key:1},[s(n(p("common.detail.none")),1)],64))]}),_:2},1024)])])):_("",!0),s(),i("div",null,[i("h3",null,n(p("policies.routes.item.config")),1),s(),i("div",N,[r(E,{resource:t.policy.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:a(({copy:o,copying:w})=>[w?(c(),m(v,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:y=>{o(h=>h(y))},onError:y=>{o((h,V)=>V(y))}},null,8,["src","onChange","onError"])):_("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]))]}),_:2},1024)]),_:1})}}}),j=P(q,[["__scopeId","data-v-be647fab"]]);export{j as default};
