import{E as m}from"./EnvoyData-64b1b1ce.js";import{a as u}from"./dataplane-dcd0858b.js";import{d as _,a as e,o as f,b as h,w as t,e as n,p as w,f as g,q as x}from"./index-d50afca2.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-2e43d83c.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-ae8a9f10.js";const q=_({__name:"DataPlaneXdsConfigView",props:{data:{}},setup(s){const o=s;return(C,V)=>{const r=e("RouteTitle"),p=e("KCard"),d=e("AppView"),c=e("RouteView");return f(),h(c,{name:"data-plane-xds-config-view",params:{mesh:"",dataPlane:"",codeSearch:""}},{default:t(({route:a,t:l})=>[n(d,null,{title:t(()=>[w("h2",null,[n(r,{title:l("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"])])]),default:t(()=>[g(),n(p,null,{body:t(()=>[n(m,{status:x(u)(o.data.dataplane,o.data.dataplaneInsight).status,resource:"Data Plane Proxy",src:`/meshes/${a.params.mesh}/dataplanes/${a.params.dataPlane}/data-path/xds`,query:a.params.codeSearch,onQueryChange:i=>a.update({codeSearch:i})},null,8,["status","src","query","onQueryChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{q as default};
