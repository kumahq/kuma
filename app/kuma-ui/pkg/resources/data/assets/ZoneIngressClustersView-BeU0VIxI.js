import{_ as p}from"./EnvoyData.vue_vue_type_script_setup_true_lang-CuQJYRMJ.js";import{d,i as s,o as m,a as u,w as n,j as a,g,k as _}from"./index-vd7wH-Zb.js";import"./kong-icons.es350-D9NAJNMW.js";import"./CodeBlock-V6yCCn_C.js";const z=d({__name:"ZoneIngressClustersView",setup(f){return(h,C)=>{const t=s("RouteTitle"),r=s("KCard"),i=s("AppView"),c=s("RouteView");return m(),u(c,{name:"zone-ingress-clusters-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:l})=>[a(i,null,{title:n(()=>[g("h2",null,[a(t,{title:l("zone-ingresses.routes.item.navigation.zone-ingress-clusters-view")},null,8,["title"])])]),default:n(()=>[_(),a(r,null,{default:n(()=>[a(p,{resource:"Zone",src:`/zone-ingresses/${e.params.zoneIngress}/data-path/clusters`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{z as default};
