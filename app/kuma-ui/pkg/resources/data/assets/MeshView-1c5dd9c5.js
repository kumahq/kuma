import{d as f,L as p,r as d,o as c,a as i,w as n,h as a,b as h,g as w,i as R}from"./index-231ca628.js";import{j as b,f as v,_ as C}from"./RouteView.vue_vue_type_script_setup_true_lang-ed887f62.js";import{N}from"./NavTabs-37b7b891.js";import"./kongponents.es-32169022.js";const j=f({__name:"MeshView",setup(V){var r;const l=b(),_=(((r=p().getRoutes().find(e=>e.name==="mesh-detail-view"))==null?void 0:r.children)??[]).map(e=>{var m,u;const s=typeof e.name>"u"?(m=e.children)==null?void 0:m[0]:e,t=s.name,o=((u=s.meta)==null?void 0:u.module)??"";return{title:l.t(`meshes.routes.item.navigation.${t}`),routeName:t,module:o}});return(e,s)=>{const t=d("RouterView");return c(),i(C,null,{default:n(()=>[a(v,null,{default:n(()=>[a(N,{class:"route-mesh-view-tabs",tabs:h(_)},null,8,["tabs"]),w(),a(t,null,{default:n(o=>[(c(),i(R(o.Component),{key:o.route.path}))]),_:1})]),_:1})]),_:1})}}});export{j as default};
