import{d as $,l as g,a2 as k,a as o,o as l,b as c,w as t,e as s,q as w,p as B,a3 as C,f as b,H as G,I as N,c as T,F as D,J as P}from"./index-d50afca2.js";import{N as q}from"./NavTabs-84db4ee8.js";const I=$({__name:"DataPlaneDetailTabsView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(h){var _;const{t:p}=g(),v=k(),n=h,x=(((_=v.getRoutes().find(a=>a.name===`${n.isGatewayView?"gateway":"data-plane"}-detail-tabs-view`))==null?void 0:_.children)??[]).map(a=>{var m,r;const d=typeof a.name>"u"?(m=a.children)==null?void 0:m[0]:a,i=d.name,u=((r=d.meta)==null?void 0:r.module)??"";return{title:p(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.navigation.${i}`),routeName:i,module:u}});return(a,d)=>{const i=o("RouteTitle"),u=o("RouterView"),f=o("DataSource"),m=o("AppView"),r=o("RouteView");return l(),c(r,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:t(({route:e})=>[s(m,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:`${n.isGatewayView?"gateway":"data-plane"}-list-view`,params:{mesh:e.params.mesh}},text:w(p)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[B("h1",null,[s(C,{text:e.params.dataPlane},{default:t(()=>[s(i,{title:w(p)(`${n.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:e.params.dataPlane})},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[b(),s(f,{src:`/meshes/${e.params.mesh}/dataplane-overviews/${e.params.dataPlane}`},{default:t(({data:y,error:V})=>[V?(l(),c(G,{key:0,error:V},null,8,["error"])):y===void 0?(l(),c(N,{key:1})):(l(),T(D,{key:2},[s(q,{class:"route-data-plane-view-tabs",tabs:w(x)},null,8,["tabs"]),b(),s(u,null,{default:t(R=>[(l(),c(P(R.Component),{data:y},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{I as default};
