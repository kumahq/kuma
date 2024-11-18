import{d as E,e as m,o as d,m as c,w as e,a as l,k as i,b as t,P as g,t as r,p as u}from"./index-C_eW3RRu.js";import{_ as S}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BajQNyhd.js";const D={class:"stack"},F={class:"stack-with-borders"},A={class:"mt-4"},T=E({__name:"HostnameGeneratorSummaryView",props:{items:{}},setup(y){const C=y;return(M,a)=>{const w=m("RouteTitle"),h=m("XAction"),x=m("DataSource"),k=m("AppView"),z=m("DataCollection"),v=m("RouteView");return d(),c(v,{name:"hostname-generator-summary-view",params:{name:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({route:n,t:p,can:R})=>[l(z,{items:C.items,predicate:o=>o.id===n.params.name},{item:e(({item:o})=>[l(k,null,{title:e(()=>[i("h2",null,[l(h,{to:{name:"hostname-generator-detail-view",params:{name:n.params.name}}},{default:e(()=>[l(w,{title:p("hostname-generators.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[a[7]||(a[7]=t()),i("div",D,[i("div",F,[o.namespace.length>0?(d(),c(g,{key:0,layout:"horizontal"},{title:e(()=>[t(r(p("hostname-generators.common.namespace")),1)]),body:e(()=>[t(r(o.namespace),1)]),_:2},1024)):u("",!0),a[3]||(a[3]=t()),R("use zones")&&o.zone?(d(),c(g,{key:1,layout:"horizontal"},{title:e(()=>[t(r(p("hostname-generators.common.zone")),1)]),body:e(()=>[l(h,{to:{name:"zone-cp-detail-view",params:{zone:o.zone}}},{default:e(()=>[t(r(o.zone),1)]),_:2},1032,["to"])]),_:2},1024)):u("",!0),a[4]||(a[4]=t()),o.spec.template?(d(),c(g,{key:2,layout:"horizontal"},{title:e(()=>[t(r(p("hostname-generators.common.template")),1)]),body:e(()=>[t(r(o.spec.template),1)]),_:2},1024)):u("",!0)]),a[6]||(a[6]=t()),i("div",null,[i("h3",null,r(p("hostname-generators.routes.item.config")),1),a[5]||(a[5]=t()),i("div",A,[l(S,{resource:o.$raw,"is-searchable":"",query:n.params.codeSearch,"is-filter-mode":n.params.codeFilter,"is-reg-exp-mode":n.params.codeRegExp,onQueryChange:s=>n.update({codeSearch:s}),onFilterModeChange:s=>n.update({codeFilter:s}),onRegExpModeChange:s=>n.update({codeRegExp:s})},{default:e(({copy:s,copying:V})=>[V?(d(),c(x,{key:0,src:`/hostname-generators/${n.params.name}/as/kubernetes?no-store`,onChange:_=>{s(f=>f(_))},onError:_=>{s((f,b)=>b(_))}},null,8,["src","onChange","onError"])):u("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])])])]),_:2},1024)]),_:2},1032,["items","predicate"])]),_:1})}}});export{T as default};
