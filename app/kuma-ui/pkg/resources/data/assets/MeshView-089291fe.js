import{d as f,D as p,r as d,o as c,a as i,b as s,i as a,e as h,h as R,j as w}from"./index-6fad8e68.js";import{k as C,g as V,_ as b}from"./RouteView.vue_vue_type_script_setup_true_lang-d7cfbc23.js";import{_ as k}from"./NavTabs.vue_vue_type_script_setup_true_lang-31298349.js";import"./kongponents.es-208ec824.js";const j=f({__name:"MeshView",setup(v){var r;const l=C(),_=(((r=p().getRoutes().find(e=>e.name==="mesh-detail-view"))==null?void 0:r.children)??[]).map(e=>{var m,u;const n=typeof e.name>"u"?(m=e.children)==null?void 0:m[0]:e,t=n.name,o=((u=n.meta)==null?void 0:u.module)??"";return{title:l.t(`meshes.routes.item.navigation.${t}`),routeName:t,module:o}});return(e,n)=>{const t=d("RouterView");return c(),i(b,null,{default:s(()=>[a(V,null,{default:s(()=>[a(k,{tabs:h(_)},null,8,["tabs"]),R(),a(t,null,{default:s(o=>[(c(),i(w(o.Component),{key:o.route.path}))]),_:1})]),_:1})]),_:1})}}});export{j as default};
