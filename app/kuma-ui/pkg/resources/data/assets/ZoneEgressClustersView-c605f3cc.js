import{E as l}from"./EnvoyData-51d7966a.js";import{d,a as t,o as m,b as u,w as s,e as r,p as g,f as _}from"./index-fa77c4eb.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-be05b1a9.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-76fc2d2e.js";import"./ErrorBlock-4851a125.js";import"./TextWithCopyButton-5090a504.js";import"./CopyButton-a6d96483.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-53706c7b.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-6411ade7.js";const B=d({__name:"ZoneEgressClustersView",setup(h){return(f,C)=>{const a=t("RouteTitle"),n=t("KCard"),p=t("AppView"),i=t("RouteView");return m(),u(i,{name:"zone-egress-clusters-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:s(({route:e,t:c})=>[r(p,null,{title:s(()=>[g("h2",null,[r(a,{title:c("zone-egresses.routes.item.navigation.zone-egress-clusters-view")},null,8,["title"])])]),default:s(()=>[_(),r(n,null,{body:s(()=>[r(l,{resource:"Zone",src:`/zone-egresses/${e.params.zoneEgress}/data-path/clusters`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
