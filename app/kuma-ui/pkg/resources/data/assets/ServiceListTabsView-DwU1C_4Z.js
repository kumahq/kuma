import{d as g,s as x,i as A,U as R,e as t,o as a,m as o,w as s,k as _,b as c,a as r,c as v,H as h,J as y,l as B,n as L,t as N,p as T,E as X}from"./index-Bo5vSFZC.js";const H={class:"stack"},M=["innerHTML"],G=g({__name:"ServiceListTabsView",props:{mesh:{}},setup(w){const i=w,l=x(),u=A();return R(()=>l.currentRoute.value.name,m=>{m==="service-list-tabs-view"&&l.replace(u("use service-insights",i.mesh)?{name:"service-list-view"}:{name:"mesh-service-list-view"})},{immediate:!0}),(m,D)=>{const f=t("XAction"),V=t("XActionGroup"),k=t("RouterView"),C=t("AppView"),b=t("RouteView");return a(),o(b,{name:"service-list-tabs-view",params:{mesh:""}},{default:s(({route:n,t:p})=>[_("div",H,[_("div",{innerHTML:p("services.routes.items.intro",{},{defaultMessage:""})},null,8,M),c(),r(C,null,{actions:s(()=>[r(V,{expanded:!0},{default:s(()=>[(a(!0),v(h,null,y(n.children,({name:e})=>{var d;return a(),v(h,{key:e},[!B(u)("use service-insights",i.mesh)&&["service-list-view","external-service-list-view"].includes(e)?T("",!0):(a(),o(f,{key:0,class:L({active:((d=n.child())==null?void 0:d.name)===e}),to:{name:e,params:{mesh:n.params.mesh}},"data-testid":`${e}-sub-tab`},{default:s(()=>[c(N(p(`services.routes.items.navigation.${e}`)),1)]),_:2},1032,["class","to","data-testid"]))],64)}),128))]),_:2},1024)]),default:s(()=>[c(),r(k,null,{default:s(({Component:e})=>[(a(),o(X(e),{mesh:i.mesh},null,8,["mesh"]))]),_:1})]),_:2},1024)])]),_:1})}}});export{G as default};
