import{_ as d}from"./EnvoyData.vue_vue_type_script_setup_true_lang-WF95C7aF.js";import{d as l,a,o as m,b as _,w as s,e as t,m as g,f as u}from"./index-Bqk11xPq.js";import"./CodeBlock-CFUAVpmU.js";const E=l({__name:"ZoneEgressStatsView",setup(f){return(h,x)=>{const n=a("RouteTitle"),r=a("KCard"),c=a("AppView"),p=a("RouteView");return m(),_(p,{name:"zone-egress-stats-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:s(({route:e,t:i})=>[t(c,null,{title:s(()=>[g("h2",null,[t(n,{title:i("zone-egresses.routes.item.navigation.zone-egress-stats-view")},null,8,["title"])])]),default:s(()=>[u(),t(r,null,{default:s(()=>[t(d,{resource:"Zone",src:`/zone-egresses/${e.params.zoneEgress}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{E as default};
