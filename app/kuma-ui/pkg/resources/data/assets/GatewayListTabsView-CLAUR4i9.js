import{A as V,h as e,o as i,a as l,w as t,j as s,g,b as y,F as b,B as A,l as v,k as c,t as C,r as R}from"./index-Dai73gmU.js";const k={};function x(B,D){const m=e("RouteTitle"),u=e("XAction"),p=e("XActionGroup"),d=e("DataCollection"),w=e("RouterView"),f=e("AppView"),h=e("RouteView");return i(),l(h,{name:"gateway-list-tabs-view",params:{mesh:""}},{default:t(({route:n,t:r})=>[s(f,null,{title:t(()=>{var a;return[g("h2",null,[s(m,{title:r(`${((a=n.child())==null?void 0:a.name)==="builtin-gateway-list-view"?"builtin":"delegated"}-gateways.routes.items.title`)},null,8,["title"])])]}),actions:t(()=>[s(d,{items:n.children,empty:!1},{default:t(({items:a})=>[s(p,null,{default:t(()=>[(i(!0),y(b,null,A(a,({name:o})=>{var _;return i(),l(u,{key:`${o}`,class:v({active:((_=n.child())==null?void 0:_.name)===o}),to:{name:o,params:{mesh:n.params.mesh}},"data-testid":`${o}-sub-tab`},{default:t(()=>[c(C(r(`gateways.routes.items.navigation.${o}`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),_:2},1032,["items"])]),default:t(()=>[c(),c(),s(w,null,{default:t(({Component:a})=>[(i(),l(R(a)))]),_:1})]),_:2},1024)]),_:1})}const X=V(k,[["render",x]]);export{X as default};
