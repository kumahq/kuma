import{d as b,r,o as m,p as d,w as e,b as c,ap as v,l as n,t as p,e as i,c as B,J as F,K as N,Q as D,q as w,_ as A}from"./index-BIN9nSPF.js";import{_ as G}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-3fFCInp0.js";const M={class:"stack-with-borders"},X={class:"mt-4"},q=b({__name:"BuiltinGatewaySummaryView",props:{items:{},routeName:{}},setup(f){const _=f;return(Q,a)=>{const h=r("XEmptyState"),C=r("RouteTitle"),x=r("XAction"),E=r("DataSource"),S=r("AppView"),V=r("RouteView");return m(),d(V,{name:_.routeName,params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:t,t:l})=>[c(v,{items:_.items,predicate:u=>u.id===t.params.gateway,find:!0},{empty:e(()=>[c(h,null,{title:e(()=>[n("h2",null,p(l("common.collection.summary.empty_title",{type:"Gateway"})),1)]),default:e(()=>[a[0]||(a[0]=i()),n("p",null,p(l("common.collection.summary.empty_message",{type:"Gateway"})),1)]),_:2},1024)]),default:e(({items:u})=>[(m(!0),B(F,null,N([u[0]],s=>(m(),d(S,{key:s.id},{title:e(()=>[n("h2",null,[c(x,{to:{name:"builtin-gateway-detail-view",params:{mesh:s.mesh,gateway:s.id}}},{default:e(()=>[c(C,{title:l("builtin-gateways.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[a[3]||(a[3]=i()),n("div",M,[s.namespace.length>0?(m(),d(D,{key:0,layout:"horizontal"},{title:e(()=>[i(p(l("gateways.routes.item.namespace")),1)]),body:e(()=>[i(p(s.namespace),1)]),_:2},1024)):w("",!0)]),a[4]||(a[4]=i()),n("div",null,[n("h3",null,p(l("gateways.routes.item.config")),1),a[2]||(a[2]=i()),n("div",X,[c(G,{resource:s.config,"is-searchable":"",query:t.params.codeSearch,"is-filter-mode":t.params.codeFilter,"is-reg-exp-mode":t.params.codeRegExp,onQueryChange:o=>t.update({codeSearch:o}),onFilterModeChange:o=>t.update({codeFilter:o}),onRegExpModeChange:o=>t.update({codeRegExp:o})},{default:e(({copy:o,copying:R})=>[R?(m(),d(E,{key:0,src:`/meshes/${t.params.mesh}/mesh-gateways/${t.params.gateway}/as/kubernetes?no-store`,onChange:y=>{o(g=>g(y))},onError:y=>{o((g,k)=>k(y))}},null,8,["src","onChange","onError"])):w("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1},8,["name"])}}}),z=A(q,[["__scopeId","data-v-05c905d6"]]);export{z as default};
