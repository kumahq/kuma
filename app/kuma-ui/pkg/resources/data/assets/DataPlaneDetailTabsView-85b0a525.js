import{d as R,g as $,a4 as k,r as o,o as l,i as c,w as t,j as s,k as w,p as B,a5 as C,n as h,E as G,x as N,l as T,F as D,q as P}from"./index-23176b1b.js";import{N as E}from"./NavTabs-4ef57897.js";const S=R({__name:"DataPlaneDetailTabsView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(b){var _;const{t:p}=$(),v=k(),n=b,x=(((_=v.getRoutes().find(a=>a.name===`${n.isGatewayView?"gateway":"data-plane"}-detail-tabs-view`))==null?void 0:_.children)??[]).map(a=>{var i,m;const d=typeof a.name>"u"?(i=a.children)==null?void 0:i[0]:a,r=d.name,u=((m=d.meta)==null?void 0:m.module)??"";return{title:p(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.navigation.${r}`),routeName:r,module:u}});return(a,d)=>{const r=o("RouteTitle"),u=o("RouterView"),f=o("DataSource"),i=o("AppView"),m=o("RouteView");return l(),c(m,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:t(({route:e})=>[s(i,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:`${n.isGatewayView?"gateway":"data-plane"}-list-view`,params:{mesh:e.params.mesh}},text:w(p)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[B("h1",null,[s(C,{text:e.params.dataPlane},{default:t(()=>[s(r,{title:w(p)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:e.params.dataPlane}),render:!0},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[h(),s(f,{src:`/meshes/${e.params.mesh}/dataplane-overviews/${e.params.dataPlane}`},{default:t(({data:y,error:V})=>[V?(l(),c(G,{key:0,error:V},null,8,["error"])):y===void 0?(l(),c(N,{key:1})):(l(),T(D,{key:2},[s(E,{class:"route-data-plane-view-tabs",tabs:w(x)},null,8,["tabs"]),h(),s(u,null,{default:t(g=>[(l(),c(P(g.Component),{data:y},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{S as default};
