import{E as l}from"./EnvoyData-b9df63cd.js";import{d as c,a as o,o as m,b as u,w as t,e as n,p as _,f as g}from"./index-baa571c4.js";import"./index-52545d1d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2bcf6524.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-218784c7.js";import"./ErrorBlock-439da12c.js";import"./TextWithCopyButton-47107f36.js";import"./CopyButton-6c8cb7cc.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-ce954803.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-b011efe4.js";const B=c({__name:"DataPlaneXdsConfigView",setup(h){return(f,x)=>{const r=o("RouteTitle"),p=o("KCard"),s=o("AppView"),d=o("RouteView");return m(),u(d,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:e,t:i})=>[n(s,null,{title:t(()=>[_("h2",null,[n(r,{title:i("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"])])]),default:t(()=>[g(),n(p,null,{body:t(()=>[n(l,{resource:"Data Plane Proxy",src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/xds`,query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter==="true","is-reg-exp-mode":e.params.codeRegExp==="true",onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["src","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
