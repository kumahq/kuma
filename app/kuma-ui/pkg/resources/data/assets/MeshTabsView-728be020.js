import{d,a2 as f,a3 as w,r as n,o as R,g as V,w as m,h as s,m as b,a4 as x,i as p,l as _}from"./index-ecc7df9d.js";import{N as v}from"./NavTabs-93c19855.js";const k=d({__name:"MeshTabsView",setup(N){var c;const{t:i}=f(),h=(((c=w().getRoutes().find(e=>e.name==="mesh-tabs-view"))==null?void 0:c.children)??[]).map(e=>{var a,t;const r=typeof e.name>"u"?(a=e.children)==null?void 0:a[0]:e,o=r.name,u=((t=r.meta)==null?void 0:t.module)??"";return{title:i(`meshes.routes.item.navigation.${o}`),routeName:o,module:u}});return(e,r)=>{const o=n("RouteTitle"),u=n("RouterView"),l=n("AppView"),a=n("RouteView");return R(),V(a,{name:"mesh-tabs-view",params:{mesh:""}},{default:m(({route:t})=>[s(l,null,{title:m(()=>[b("h1",null,[s(x,{text:t.params.mesh},{default:m(()=>[s(o,{title:p(i)("meshes.routes.item.title",{name:t.params.mesh}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:m(()=>[_(),s(v,{class:"route-mesh-view-tabs",tabs:p(h)},null,8,["tabs"]),_(),s(u)]),_:2},1024)]),_:1})}}});export{k as default};
