import{d as V,a as l,o as i,b as d,w as a,e as m,m as s,f as t,t as p,c as _,X as b,p as u,F as E,v as F,x as P,Q as B,_ as T}from"./index-cf10d15e.js";import{_ as $}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-dae16a2a.js";import{_ as D}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-51f56ed5.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-4bf3aea6.js";import"./uniqueId-90cc9b93.js";import"./CopyButton-0aa5d830.js";import"./index-fce48c05.js";import"./toYaml-4e00099e.js";const I=r=>(F("data-v-e6c20b15"),r=r(),P(),r),M={class:"summary-title-wrapper"},N=I(()=>s("img",{"aria-hidden":"true",src:B},null,-1)),Q={class:"summary-title"},q={key:1,class:"stack"},A={key:0},K={class:"mt-4 stack"},L={key:0},X={class:"mt-4"},j=V({__name:"PolicySummaryView",props:{policy:{default:void 0},policyType:{}},setup(r){const c=r;return(z,G)=>{const f=l("RouteTitle"),v=l("RouterLink"),R=l("KBadge"),k=l("DataSource"),w=l("AppView"),x=l("RouteView");return i(),d(x,{name:"policy-summary-view",params:{mesh:"",policyPath:"",policy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:n})=>[m(w,null,{title:a(()=>[s("div",M,[N,t(),s("h2",Q,[m(v,{to:{name:"policy-detail-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath,policy:e.params.policy}}},{default:a(()=>[m(f,{title:n("policies.routes.item.title",{name:e.params.policy})},null,8,["title"])]),_:2},1032,["to"])])])]),default:a(()=>{var h;return[t(),c.policy===void 0?(i(),d($,{key:0},{message:a(()=>[s("p",null,p(n("common.collection.summary.empty_message",{type:c.policyType.name})),1)]),default:a(()=>[t(p(n("common.collection.summary.empty_title",{type:c.policyType.name}))+" ",1)]),_:2},1024)):(i(),_("div",q,[(h=c.policy.spec)!=null&&h.targetRef?(i(),_("div",A,[s("h3",null,p(n("policies.routes.item.overview")),1),t(),s("div",K,[m(b,null,{title:a(()=>[t(p(n("http.api.property.targetRef")),1)]),body:a(()=>{var o;return[(o=c.policy.spec)!=null&&o.targetRef?(i(),d(R,{key:0,appearance:"neutral"},{default:a(()=>[t(p(c.policy.spec.targetRef.kind),1),c.policy.spec.targetRef.name?(i(),_("span",L,[t(":"),s("b",null,p(c.policy.spec.targetRef.name),1)])):u("",!0)]),_:1})):(i(),_(E,{key:1},[t(p(n("common.detail.none")),1)],64))]}),_:2},1024)])])):u("",!0),t(),s("div",null,[s("h3",null,p(n("policies.routes.item.config")),1),t(),s("div",X,[m(D,{resource:c.policy.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:a(({copy:o,copying:C})=>[C?(i(),d(k,{key:0,src:`/meshes/${e.params.mesh}/policy-path/${e.params.policyPath}/policy/${e.params.policy}/as/kubernetes?no-store`,onChange:y=>{o(g=>g(y))},onError:y=>{o((g,S)=>S(y))}},null,8,["src","onChange","onError"])):u("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]))]}),_:2},1024)]),_:1})}}});const oe=T(j,[["__scopeId","data-v-e6c20b15"]]);export{oe as default};
