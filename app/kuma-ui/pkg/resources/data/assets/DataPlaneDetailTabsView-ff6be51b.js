import{d as $,l as g,W as k,a as n,o as m,b as p,w as t,e as s,q as w,p as B,f as b,c as T,F as C,D}from"./index-7a0947c2.js";import{E as G}from"./ErrorBlock-78880c60.js";import{_ as N}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-a6d76488.js";import{N as P}from"./NavTabs-e9c664ed.js";import{T as E}from"./TextWithCopyButton-3aa03737.js";import"./index-fce48c05.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-1c689249.js";import"./CopyButton-a5c25cdd.js";const J=$({__name:"DataPlaneDetailTabsView",props:{isGatewayView:{type:Boolean,required:!1,default:!1}},setup(h){var _;const{t:c}=g(),v=k(),o=h,x=(((_=v.getRoutes().find(a=>a.name===`${o.isGatewayView?"gateway":"data-plane"}-detail-tabs-view`))==null?void 0:_.children)??[]).map(a=>{var i,l;const d=typeof a.name>"u"?(i=a.children)==null?void 0:i[0]:a,r=d.name,u=((l=d.meta)==null?void 0:l.module)??"";return{title:c(`${o.isGatewayView?"gateways":"data-planes"}.routes.item.navigation.${r}`),routeName:r,module:u}});return(a,d)=>{const r=n("RouteTitle"),u=n("RouterView"),f=n("DataSource"),i=n("AppView"),l=n("RouteView");return m(),p(l,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:t(({route:e})=>[s(i,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:`${o.isGatewayView?"gateway":"data-plane"}-list-view`,params:{mesh:e.params.mesh}},text:w(c)(`${o.isGatewayView?"gateways":"data-planes"}.routes.item.breadcrumbs`)}]},{title:t(()=>[B("h1",null,[s(E,{text:e.params.dataPlane},{default:t(()=>[s(r,{title:w(c)(`${o.isGatewayView?"gateways":"data-planes"}.routes.item.title`,{name:e.params.dataPlane})},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[b(),s(f,{src:`/meshes/${e.params.mesh}/dataplane-overviews/${e.params.dataPlane}`},{default:t(({data:y,error:V})=>[V?(m(),p(G,{key:0,error:V},null,8,["error"])):y===void 0?(m(),p(N,{key:1})):(m(),T(C,{key:2},[s(P,{class:"route-data-plane-view-tabs",tabs:w(x)},null,8,["tabs"]),b(),s(u,null,{default:t(R=>[(m(),p(D(R.Component),{data:y},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{J as default};
