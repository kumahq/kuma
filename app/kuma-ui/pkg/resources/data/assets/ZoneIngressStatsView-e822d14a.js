import{E as m}from"./EnvoyData-2bd5eaaf.js";import{d,a as t,o as l,b as g,w as a,e as n,m as u,f as _}from"./index-a963f507.js";import"./index-fce48c05.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-3088bbc1.js";import"./uniqueId-90cc9b93.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-3297bc03.js";import"./ErrorBlock-a09a5c02.js";import"./TextWithCopyButton-442c5ee6.js";import"./CopyButton-003985ad.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-0ab8cf0c.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-c187db3a.js";const B=d({__name:"ZoneIngressStatsView",setup(f){return(h,x)=>{const s=t("RouteTitle"),r=t("KCard"),i=t("AppView"),p=t("RouteView");return l(),g(p,{name:"zone-ingress-stats-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c})=>[n(i,null,{title:a(()=>[u("h2",null,[n(s,{title:c("zone-ingresses.routes.item.navigation.zone-ingress-stats-view")},null,8,["title"])])]),default:a(()=>[_(),n(r,null,{default:a(()=>[n(m,{resource:"Zone",src:`/zone-ingresses/${e.params.zoneIngress}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
