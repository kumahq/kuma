import{E as l}from"./EnvoyData-6e4ba4d3.js";import{d,a as t,o as m,b as u,w as s,e as r,p as g,f as _}from"./index-f56c27ab.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-68115bc8.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-2656c87a.js";import"./ErrorBlock-aa131c0d.js";import"./TextWithCopyButton-6a682cf8.js";import"./CopyButton-7b31a54c.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-af2f8e99.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-b984e5b1.js";const B=d({__name:"ZoneEgressClustersView",setup(h){return(f,C)=>{const a=t("RouteTitle"),n=t("KCard"),p=t("AppView"),i=t("RouteView");return m(),u(i,{name:"zone-egress-clusters-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:s(({route:e,t:c})=>[r(p,null,{title:s(()=>[g("h2",null,[r(a,{title:c("zone-egresses.routes.item.navigation.zone-egress-clusters-view")},null,8,["title"])])]),default:s(()=>[_(),r(n,null,{body:s(()=>[r(l,{resource:"Zone",src:`/zone-egresses/${e.params.zoneEgress}/data-path/clusters`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
