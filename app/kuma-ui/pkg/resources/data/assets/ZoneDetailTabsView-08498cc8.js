import{d as b,L as z,a8 as v,r as w,o,g as i,w as t,h as a,A as x,i as d,m as h,a9 as k,C,l as _,p as V,E as y,s as B,j as N,F as $,n as R,_ as g}from"./index-0ab7ff60.js";import{N as T}from"./NavTabs-bbfe4fd0.js";const I=b({__name:"ZoneDetailTabsView",setup(E){var p;const{t:l}=z(),f=(((p=v().getRoutes().find(e=>e.name==="zone-cp-detail-tabs-view"))==null?void 0:p.children)??[]).map(e=>{var n,c;const m=typeof e.name>"u"?(n=e.children)==null?void 0:n[0]:e,r=m.name,s=((c=m.meta)==null?void 0:c.module)??"";return{title:l(`zone-cps.routes.item.navigation.${r}`),routeName:r,module:s}});return(e,m)=>{const r=w("RouterView");return o(),i(g,{name:"zone-cp-detail-tabs-view","data-testid":"zone-cp-detail-tabs-view"},{default:t(({route:s})=>[a(x,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:d(l)("zone-cps.routes.item.breadcrumbs")}]},{title:t(()=>[h("h1",null,[a(k,{text:s.params.zone},{default:t(()=>[a(C,{title:d(l)("zone-cps.routes.item.title",{name:s.params.zone}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[_(),a(V,{src:`/zone-cps/${s.params.zone}`},{default:t(({data:u,error:n})=>[n!==void 0?(o(),i(y,{key:0,error:n},null,8,["error"])):u===void 0?(o(),i(B,{key:1})):(o(),N($,{key:2},[a(T,{class:"route-zone-detail-view-tabs",tabs:d(f)},null,8,["tabs"]),_(),a(r,null,{default:t(c=>[(o(),i(R(c.Component),{data:u},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{I as default};
