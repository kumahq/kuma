import{q as b,e as t,o as l,m as d,w as a,a as o,b as c,k as p,c as k,J as x,K as A,n as C,t as R,F as T}from"./index-CKcsX_-l.js";const h={},$={class:"stack"},B=["innerHTML"];function D(L,s){const _=t("RouteTitle"),w=t("XAction"),f=t("XActionGroup"),g=t("DataCollection"),y=t("RouterView"),V=t("AppView"),v=t("RouteView");return l(),d(v,{name:"gateway-list-tabs-view",params:{mesh:""}},{default:a(({route:n,t:r})=>{var m;return[o(_,{render:!1,title:r(`${((m=n.child())==null?void 0:m.name)==="builtin-gateway-list-view"?"builtin":"delegated"}-gateways.routes.items.title`)},null,8,["title"]),s[2]||(s[2]=c()),p("div",$,[o(V,null,{actions:a(()=>[o(g,{items:n.children,empty:!1},{default:a(({items:i})=>[o(f,{expanded:!0},{default:a(()=>[(l(!0),k(x,null,A(i,({name:e})=>{var u;return l(),d(w,{key:`${e}`,class:C({active:((u=n.child())==null?void 0:u.name)===e}),to:{name:e,params:{mesh:n.params.mesh}},"data-testid":`${e}-sub-tab`},{default:a(()=>[c(R(r(`gateways.routes.items.navigation.${e}.label`)),1)]),_:2},1032,["class","to","data-testid"])}),128))]),_:2},1024)]),_:2},1032,["items"])]),default:a(()=>{var i;return[s[0]||(s[0]=c()),p("div",{innerHTML:r(`gateways.routes.items.navigation.${(i=n.child())==null?void 0:i.name}.description`,{},{defaultMessage:""})},null,8,B),s[1]||(s[1]=c()),o(y,null,{default:a(({Component:e})=>[(l(),d(T(e)))]),_:1})]}),_:2},1024)])]}),_:1})}const G=b(h,[["render",D]]);export{G as default};
