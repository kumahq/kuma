import{_ as c}from"./EnvoyData.vue_vue_type_script_setup_true_lang-BsZo_vtP.js";import{d as i,r as o,o as m,m as _,w as n,b as t,e as f}from"./index-BRBWbknO.js";import"./kong-icons.es350-BcALoRcG.js";import"./CodeBlock-BW7MdVHn.js";const V=i({__name:"DataPlaneXdsConfigView",setup(g){return(u,h)=>{const s=o("RouteTitle"),d=o("KCard"),r=o("AppView"),p=o("RouteView");return m(),_(p,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:l})=>[t(s,{render:!1,title:l("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"]),f(),t(r,null,{default:n(()=>[t(d,null,{default:n(()=>[t(c,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/xds`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{V as default};
