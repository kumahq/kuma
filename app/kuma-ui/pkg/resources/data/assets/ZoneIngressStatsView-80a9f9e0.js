import{E as d}from"./EnvoyData-b9df63cd.js";import{d as m,a as t,o as l,b as g,w as a,e as n,p as u,f as _}from"./index-baa571c4.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2bcf6524.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-218784c7.js";import"./ErrorBlock-439da12c.js";import"./TextWithCopyButton-47107f36.js";import"./CopyButton-6c8cb7cc.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-ce954803.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-b011efe4.js";const S=m({__name:"ZoneIngressStatsView",setup(h){return(f,x)=>{const s=t("RouteTitle"),r=t("KCard"),i=t("AppView"),p=t("RouteView");return l(),g(p,{name:"zone-ingress-stats-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:c})=>[n(i,null,{title:a(()=>[u("h2",null,[n(s,{title:c("zone-ingresses.routes.item.navigation.zone-ingress-stats-view")},null,8,["title"])])]),default:a(()=>[_(),n(r,null,{body:a(()=>[n(d,{resource:"Zone",src:`/zone-ingresses/${e.params.zoneIngress}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{S as default};
