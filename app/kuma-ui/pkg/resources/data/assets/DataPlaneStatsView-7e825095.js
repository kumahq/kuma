import{E as l}from"./EnvoyData-b9df63cd.js";import{d as m,a as t,o as c,b as u,w as o,e as n,p as _,f as h}from"./index-baa571c4.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2bcf6524.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-218784c7.js";import"./ErrorBlock-439da12c.js";import"./TextWithCopyButton-47107f36.js";import"./CopyButton-6c8cb7cc.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-ce954803.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-b011efe4.js";const S=m({__name:"DataPlaneStatsView",setup(g){return(f,x)=>{const s=t("RouteTitle"),r=t("KCard"),p=t("AppView"),i=t("RouteView");return c(),u(i,{name:"data-plane-stats-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:d})=>[n(p,null,{title:o(()=>[_("h2",null,[n(s,{title:d("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"])])]),default:o(()=>[h(),n(r,null,{body:o(()=>[n(l,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/stats`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{S as default};
