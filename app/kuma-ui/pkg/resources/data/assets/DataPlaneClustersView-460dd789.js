import{_ as c}from"./EnvoyData.vue_vue_type_script_setup_true_lang-95647c86.js";import{d,a as t,o as m,b as u,w as o,e as n,m as _,f as h}from"./index-3ddd0e9e.js";import"./index-fce48c05.js";import"./CodeBlock-56f65f8d.js";import"./uniqueId-90cc9b93.js";import"./ErrorBlock-2be9cd06.js";import"./TextWithCopyButton-4870eafb.js";import"./CopyButton-f0ea0e69.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-39d02562.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-4407ccd5.js";const B=d({__name:"DataPlaneClustersView",setup(f){return(g,x)=>{const s=t("RouteTitle"),r=t("KCard"),p=t("AppView"),l=t("RouteView");return m(),u(l,{name:"data-plane-clusters-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:i})=>[n(p,null,{title:o(()=>[_("h2",null,[n(s,{title:i("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"])])]),default:o(()=>[h(),n(r,null,{default:o(()=>[n(c,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
