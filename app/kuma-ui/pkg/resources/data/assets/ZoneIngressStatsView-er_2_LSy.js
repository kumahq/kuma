import{_ as d}from"./EnvoyData.vue_vue_type_script_setup_true_lang-ZCzPR866.js";import{d as l,a,o as m,b as _,w as n,e as s,m as g,f as u}from"./index-pAyRVwwQ.js";import"./CodeBlock-6c7dCnil.js";const V=l({__name:"ZoneIngressStatsView",setup(f){return(h,x)=>{const t=a("RouteTitle"),r=a("KCard"),i=a("AppView"),c=a("RouteView");return m(),_(c,{name:"zone-ingress-stats-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:p})=>[s(i,null,{title:n(()=>[g("h2",null,[s(t,{title:p("zone-ingresses.routes.item.navigation.zone-ingress-stats-view")},null,8,["title"])])]),default:n(()=>[u(),s(r,null,{default:n(()=>[s(d,{resource:"Zone",src:`/zone-ingresses/${e.params.zoneIngress}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{V as default};
