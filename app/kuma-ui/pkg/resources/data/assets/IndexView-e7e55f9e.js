import{d as V,g as R,a4 as k,r as o,o as n,i as u,w as t,j as s,k as l,p as E,a5 as h,n as w,E as y,x as B,l as C,F as N,q as T}from"./index-21079cd9.js";import{N as D}from"./NavTabs-ddecd1ce.js";const j=V({__name:"IndexView",setup($){var _;const{t:a}=R(),z=(((_=k().getRoutes().find(e=>e.name==="zone-egress-detail-tabs-view"))==null?void 0:_.children)??[]).map(e=>{var c,i;const m=typeof e.name>"u"?(c=e.children)==null?void 0:c[0]:e,r=m.name,p=((i=m.meta)==null?void 0:i.module)??"";return{title:a(`zone-egresses.routes.item.navigation.${r}`),routeName:r,module:p}});return(e,m)=>{const r=o("RouteTitle"),p=o("RouterView"),f=o("DataSource"),c=o("AppView"),i=o("RouteView");return n(),u(i,{name:"zone-egress-detail-tabs-view",params:{zoneEgress:""}},{default:t(({route:d,can:v})=>[s(c,{breadcrumbs:[...v("use zones")?[{to:{name:"zone-cp-list-view"},text:l(a)("zone-cps.routes.item.breadcrumbs")}]:[],{to:{name:"zone-egress-list-view"},text:l(a)("zone-egresses.routes.item.breadcrumbs")}]},{title:t(()=>[E("h1",null,[s(h,{text:d.params.zoneEgress},{default:t(()=>[s(r,{title:l(a)("zone-egresses.routes.item.title",{name:d.params.zoneEgress}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[w(),s(f,{src:`/zone-egress-overviews/${d.params.zoneEgress}`},{default:t(({data:g,error:b})=>[b!==void 0?(n(),u(y,{key:0,error:b},null,8,["error"])):g===void 0?(n(),u(B,{key:1})):(n(),C(N,{key:2},[s(D,{class:"route-zone-egress-detail-view-tabs",tabs:l(z)},null,8,["tabs"]),w(),s(p,null,{default:t(x=>[(n(),u(T(x.Component),{data:g},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{j as default};
