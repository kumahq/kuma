import{d as h,I as b,r as w,o as c,a as p,w as s,h as o,q as R,b as _,g as f,i as V}from"./index-30c3bdbc.js";import{g as v,A as N,_ as C}from"./RouteView.vue_vue_type_script_setup_true_lang-1dd6f4c1.js";import{_ as g}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cbf5001a.js";import{N as x}from"./NavTabs-2c5e4459.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b3f1f8ad.js";const D=h({__name:"MeshTabsView",setup(k){var u;const{t:m}=v(),d=(((u=b().getRoutes().find(e=>e.name==="mesh-tabs-view"))==null?void 0:u.children)??[]).map(e=>{var i,l;const a=typeof e.name>"u"?(i=e.children)==null?void 0:i[0]:e,t=a.name,n=((l=a.meta)==null?void 0:l.module)??"";return{title:m(`meshes.routes.item.navigation.${t}`),routeName:t,module:n}});return(e,a)=>{const t=w("RouterView");return c(),p(C,null,{default:s(({route:n})=>[o(N,null,{title:s(()=>[R("h1",null,[o(g,{title:_(m)("meshes.routes.item.title",{name:n.params.mesh}),render:!0},null,8,["title"])])]),default:s(()=>[f(),o(x,{class:"route-mesh-view-tabs",tabs:_(d)},null,8,["tabs"]),f(),o(t,null,{default:s(r=>[(c(),p(V(r.Component),{key:r.route.path}))]),_:2},1024)]),_:2},1024)]),_:1})}}});export{D as default};
