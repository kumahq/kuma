import{d as b,e as r,o as i,m as p,w as e,a as c,k as m,b as a,R as u,t as s,p as d}from"./index-COT-_p62.js";import{_ as E}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CJpEpjzv.js";const S={class:"stack"},D={class:"stack-with-borders"},F={class:"mt-4"},T=b({__name:"HostnameGeneratorSummaryView",props:{items:{}},setup(f){const y=f;return(A,M)=>{const C=r("RouteTitle"),h=r("XAction"),w=r("DataSource"),x=r("AppView"),k=r("DataCollection"),z=r("RouteView");return i(),p(z,{name:"hostname-generator-summary-view",params:{name:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:t,t:l,can:R})=>[c(k,{items:y.items,predicate:o=>o.id===t.params.name},{item:e(({item:o})=>[c(x,null,{title:e(()=>[m("h2",null,[c(h,{to:{name:"hostname-generator-detail-view",params:{name:t.params.name}}},{default:e(()=>[c(C,{title:l("hostname-generators.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[a(),m("div",S,[m("div",D,[o.namespace.length>0?(i(),p(u,{key:0,layout:"horizontal"},{title:e(()=>[a(s(l("hostname-generators.common.namespace")),1)]),body:e(()=>[a(s(o.namespace),1)]),_:2},1024)):d("",!0),a(),R("use zones")&&o.zone?(i(),p(u,{key:1,layout:"horizontal"},{title:e(()=>[a(s(l("hostname-generators.common.zone")),1)]),body:e(()=>[c(h,{to:{name:"zone-cp-detail-view",params:{zone:o.zone}}},{default:e(()=>[a(s(o.zone),1)]),_:2},1032,["to"])]),_:2},1024)):d("",!0),a(),o.spec.template?(i(),p(u,{key:2,layout:"horizontal"},{title:e(()=>[a(s(l("hostname-generators.common.template")),1)]),body:e(()=>[a(s(o.spec.template),1)]),_:2},1024)):d("",!0)]),a(),m("div",null,[m("h3",null,s(l("hostname-generators.routes.item.config")),1),a(),m("div",F,[c(E,{resource:o.$raw,"is-searchable":"",query:t.params.codeSearch,"is-filter-mode":t.params.codeFilter,"is-reg-exp-mode":t.params.codeRegExp,onQueryChange:n=>t.update({codeSearch:n}),onFilterModeChange:n=>t.update({codeFilter:n}),onRegExpModeChange:n=>t.update({codeRegExp:n})},{default:e(({copy:n,copying:v})=>[v?(i(),p(w,{key:0,src:`/hostname-generators/${t.params.name}/as/kubernetes?no-store`,onChange:_=>{n(g=>g(_))},onError:_=>{n((g,V)=>V(_))}},null,8,["src","onChange","onError"])):d("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{T as default};
