import{q as V,r as e,o as n,m as r,w as t,b as s,e as i,k as m,c as v,F as b,s as k,n as x,t as A,E as C}from"./index-DxrN05KS.js";const R={},T={class:"stack"},B=["innerHTML"];function D(L,X){const u=e("RouteTitle"),p=e("XAction"),w=e("XActionGroup"),f=e("DataCollection"),g=e("RouterView"),h=e("AppView"),y=e("RouteView");return n(),r(y,{name:"gateway-list-tabs-view",params:{mesh:""}},{default:t(({route:o,t:c})=>{var _;return[s(u,{render:!1,title:c(`${((_=o.child())==null?void 0:_.name)==="builtin-gateway-list-view"?"builtin":"delegated"}-gateways.routes.items.title`)},null,8,["title"]),i(),m("div",T,[m("div",{innerHTML:c("gateways.routes.items.intro",{},{defaultMessage:""})},null,8,B),i(),s(h,null,{actions:t(()=>[s(f,{items:o.children,empty:!1},{default:t(({items:l})=>[s(w,{expanded:!0},{default:t(()=>[(n(!0),v(b,null,k(l,({name:a})=>{var d;return n(),r(p,{key:`${a}`,class:x({active:((d=o.child())==null?void 0:d.name)===a}),to:{name:a,params:{mesh:o.params.mesh}},"data-testid":`${a}-sub-tab`},{default:t(()=>[i(A(c(`gateways.routes.items.navigation.${a}`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),_:2},1032,["items"])]),default:t(()=>[i(),s(g,null,{default:t(({Component:l})=>[(n(),r(C(l)))]),_:1})]),_:2},1024)])]}),_:1})}const G=V(R,[["render",D]]);export{G as default};
