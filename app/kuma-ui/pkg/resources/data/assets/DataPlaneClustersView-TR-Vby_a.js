import{_ as d}from"./EnvoyData.vue_vue_type_script_setup_true_lang-JcZhd3BS.js";import{d as i,a as t,o as m,b as u,w as o,e as n,m as _,f as h}from"./index-KQusT94Q.js";import"./CodeBlock-A5fzLugD.js";const V=i({__name:"DataPlaneClustersView",setup(f){return(g,x)=>{const s=t("RouteTitle"),l=t("KCard"),r=t("AppView"),p=t("RouteView");return m(),u(p,{name:"data-plane-clusters-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:c})=>[n(r,null,{title:o(()=>[_("h2",null,[n(s,{title:c("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"])])]),default:o(()=>[h(),n(l,null,{default:o(()=>[n(d,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{V as default};
