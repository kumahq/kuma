import{L as V}from"./LinkBox-rX4fAdWz.js";import{d as g,a as t,o as i,b as m,w as e,e as s,m as y,F as R,n as b,t as h,f as l,H as k,c as v}from"./index-KOnKkPpw.js";const $=g({__name:"GatewayListTabsView",setup(C){return(L,x)=>{const u=t("RouteTitle"),_=t("RouterLink"),p=t("DataCollection"),d=t("RouterView"),w=t("AppView"),f=t("RouteView");return i(),m(f,{name:"gateway-list-tabs-view",params:{mesh:""}},{default:e(({route:o,t:c})=>[s(w,null,{title:e(()=>{var n;return[y("h2",null,[s(u,{title:c(`${((n=o.active)==null?void 0:n.name)==="builtin-gateway-list-view"?"builtin":"delegated"}-gateways.routes.items.title`)},null,8,["title"])])]}),actions:e(()=>[s(p,{items:o.children,empty:!1},{default:e(({items:n})=>[s(V,null,{default:e(()=>[(i(!0),v(R,null,k(n,({name:a})=>{var r;return i(),m(_,{key:`${a}`,class:b({active:((r=o.active)==null?void 0:r.name)===a}),to:{name:a,params:{mesh:o.params.mesh}},"data-testid":`${a}-sub-tab`},{default:e(()=>[l(h(c(`gateways.routes.items.navigation.${a}`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),_:2},1032,["items"])]),default:e(()=>[l(),l(),s(d)]),_:2},1024)]),_:1})}}});export{$ as default};
