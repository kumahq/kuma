import{o as v}from"./kongponents.es-dc880404.js";import{d as g,L as w,v as _,o as f,c as p,w as n,a as u,u as s,x as R,k,b as d,t as b,l as x}from"./index-271b6183.js";import{g as C,f as S,e as V}from"./RouteView.vue_vue_type_script_setup_true_lang-fadf0571.js";const N=g({__name:"MeshView",setup($){var c;const{t:m}=C(),l=w(),o=(((c=l.getRoutes().find(t=>t.name==="mesh-abstract-view"))==null?void 0:c.children)??[]).map(t=>{var r;if(typeof t.name>"u"){const a=(r=t.children)==null?void 0:r[0],e=String(a==null?void 0:a.name);return{title:m(`meshes.routes.item.navigation.${e}`),hash:e}}const i=String(t.name);return{title:m(`meshes.routes.item.navigation.${i}`),hash:i}});return(t,i)=>{const r=_("router-link"),a=_("RouterView");return f(),p(V,null,{default:n(()=>[u(S,null,{default:n(()=>[u(s(v),{class:"route-mesh-view-tabs",tabs:s(o),"has-panels":!1,"model-value":(s(o).find(e=>{var h;return(((h=s(l).currentRoute)==null?void 0:h.value.name)??"").toString().startsWith(e.hash)})??s(o)[0]).hash},R({_:2},[k(s(o),e=>({name:`${e.hash}-anchor`,fn:n(()=>[u(r,{to:{name:e.hash}},{default:n(()=>[d(b(e.title),1)]),_:2},1032,["to"])])}))]),1032,["tabs","model-value"]),d(),u(a,null,{default:n(e=>[(f(),p(x(e.Component),{key:e.route.path}))]),_:1})]),_:1})]),_:1})}}});export{N as default};
