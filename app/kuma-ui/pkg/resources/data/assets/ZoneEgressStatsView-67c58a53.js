import{E as d}from"./EnvoyData-51d7966a.js";import{d as m,a as t,o as l,b as g,w as a,e as s,p as u,f as _}from"./index-fa77c4eb.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-be05b1a9.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-76fc2d2e.js";import"./ErrorBlock-4851a125.js";import"./TextWithCopyButton-5090a504.js";import"./CopyButton-a6d96483.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-53706c7b.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-6411ade7.js";const S=m({__name:"ZoneEgressStatsView",setup(h){return(f,x)=>{const r=t("RouteTitle"),n=t("KCard"),p=t("AppView"),i=t("RouteView");return l(),g(i,{name:"zone-egress-stats-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c})=>[s(p,null,{title:a(()=>[u("h2",null,[s(r,{title:c("zone-egresses.routes.item.navigation.zone-egress-stats-view")},null,8,["title"])])]),default:a(()=>[_(),s(n,null,{body:a(()=>[s(d,{resource:"Zone",src:`/zone-egresses/${e.params.zoneEgress}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{S as default};
