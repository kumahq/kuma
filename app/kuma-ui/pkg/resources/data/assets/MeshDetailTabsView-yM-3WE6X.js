import{d as h,l as f,U as w,a as m,o as R,b,w as i,q as p,e as o,m as V,f as d}from"./index-H9kuPi5I.js";import{N as T}from"./NavTabs-9SOFDSU2.js";import{T as v}from"./TextWithCopyButton-WiPvXg9-.js";import"./CopyButton-6xMoQ2pP.js";import"./index-FZCiQto1.js";const A=h({__name:"MeshDetailTabsView",setup(x){var u;const{t:l}=f(),_=(((u=w().getRoutes().find(e=>e.name==="mesh-detail-tabs-view"))==null?void 0:u.children)??[]).filter(e=>{var t;return!((t=e.meta)!=null&&t.shouldIgnoreInNavTabs)}).map(e=>{var n,s;const t=typeof e.name>"u"?(n=e.children)==null?void 0:n[0]:e,a=t.name,r=((s=t.meta)==null?void 0:s.module)??"";return{title:l(`meshes.routes.item.navigation.${a}`),routeName:a,module:r}});return(e,t)=>{const a=m("RouteTitle"),r=m("RouterView"),c=m("AppView"),n=m("RouteView");return R(),b(n,{name:"mesh-detail-tabs-view",params:{mesh:""}},{default:i(({route:s})=>[o(c,null,{title:i(()=>[V("h1",null,[o(v,{text:s.params.mesh},{default:i(()=>[o(a,{title:p(l)("meshes.routes.item.title",{name:s.params.mesh})},null,8,["title"])]),_:2},1032,["text"])])]),default:i(()=>[d(),o(T,{class:"route-mesh-view-tabs",tabs:p(_)},null,8,["tabs"]),d(),o(r)]),_:2},1024)]),_:1})}}});export{A as default};
