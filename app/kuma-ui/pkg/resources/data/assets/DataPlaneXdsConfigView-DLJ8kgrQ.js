import{_ as i}from"./EnvoyData.vue_vue_type_script_setup_true_lang-CuQJYRMJ.js";import{d as c,i as o,o as m,a as _,w as t,j as n,g,k as u}from"./index-vd7wH-Zb.js";import"./kong-icons.es350-D9NAJNMW.js";import"./CodeBlock-V6yCCn_C.js";const y=c({__name:"DataPlaneXdsConfigView",setup(f){return(h,x)=>{const s=o("RouteTitle"),d=o("KCard"),p=o("AppView"),r=o("RouteView");return m(),_(r,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:e,t:l})=>[n(p,null,{title:t(()=>[g("h2",null,[n(s,{title:l("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"])])]),default:t(()=>[u(),n(d,null,{default:t(()=>[n(i,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/xds`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{y as default};
