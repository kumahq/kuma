import{_ as c}from"./EnvoyData.vue_vue_type_script_setup_true_lang-WF95C7aF.js";import{d as i,a as o,o as m,b as _,w as t,e as n,m as u,f}from"./index-Bqk11xPq.js";import"./CodeBlock-CFUAVpmU.js";const V=i({__name:"DataPlaneXdsConfigView",setup(g){return(h,x)=>{const s=o("RouteTitle"),d=o("KCard"),l=o("AppView"),p=o("RouteView");return m(),_(p,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:e,t:r})=>[n(l,null,{title:t(()=>[u("h2",null,[n(s,{title:r("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"])])]),default:t(()=>[f(),n(d,null,{default:t(()=>[n(c,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/xds`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{V as default};
