import{_ as v,r as t,o as l,q as r,w as a,b as n,e as c,m as h,c as x,K as A,L as C,n as R,t as X,G as k}from"./index-U3igbuyl.js";const $={},B={class:"stack"};function D(G,s){const u=t("RouteTitle"),_=t("XAction"),w=t("XActionGroup"),f=t("DataCollection"),g=t("XI18n"),y=t("RouterView"),V=t("AppView"),b=t("RouteView");return l(),r(b,{name:"gateway-list-tabs-view",params:{mesh:""}},{default:a(({route:o,t:m})=>{var d;return[n(u,{render:!1,title:m(`${((d=o.child())==null?void 0:d.name)==="builtin-gateway-list-view"?"builtin":"delegated"}-gateways.routes.items.title`)},null,8,["title"]),s[2]||(s[2]=c()),h("div",B,[n(V,null,{actions:a(()=>[n(f,{items:o.children,empty:!1},{default:a(({items:i})=>[n(w,{expanded:!0},{default:a(()=>[(l(!0),x(A,null,C(i,({name:e})=>{var p;return l(),r(_,{key:`${e}`,class:R({active:((p=o.child())==null?void 0:p.name)===e}),to:{name:e,params:{mesh:o.params.mesh}},"data-testid":`${e}-sub-tab`},{default:a(()=>[c(X(m(`gateways.routes.items.navigation.${e}.label`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),_:2},1032,["items"])]),default:a(()=>{var i;return[s[0]||(s[0]=c()),n(g,{path:`gateways.routes.items.navigation.${(i=o.child())==null?void 0:i.name}.description`,"default-message":""},null,8,["path"]),s[1]||(s[1]=c()),n(y,null,{default:a(({Component:e})=>[(l(),r(k(e)))]),_:1})]}),_:2},1024)])]}),_:1})}const L=v($,[["render",D]]);export{L as default};
