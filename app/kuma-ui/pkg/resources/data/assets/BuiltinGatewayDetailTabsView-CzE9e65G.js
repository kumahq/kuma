import{d as V,e as t,o as i,m as l,w as e,a,l as x,ao as D,c as v,$ as R,p as T,b as c,Q as C,J as k,t as A,E as B}from"./index-DS1MbHFW.js";const S={key:0},E=V({__name:"BuiltinGatewayDetailTabsView",setup(X){return($,L)=>{const _=t("RouteTitle"),u=t("XAction"),p=t("XTabs"),d=t("RouterView"),w=t("DataLoader"),h=t("AppView"),b=t("DataSource"),f=t("RouteView");return i(),l(f,{name:"builtin-gateway-detail-tabs-view",params:{mesh:"",gateway:""}},{default:e(({route:s,t:m,uri:y})=>[a(b,{src:y(x(D),"/meshes/:mesh/mesh-gateways/:name",{mesh:s.params.mesh,name:s.params.gateway})},{default:e(({data:o,error:g})=>[a(h,{docs:m("builtin-gateways.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"builtin-gateway-list-view",params:{mesh:s.params.mesh}},text:m("builtin-gateways.routes.item.breadcrumbs")}]},{title:e(()=>[o?(i(),v("h1",S,[a(R,{text:o.name},{default:e(()=>[a(_,{title:m("builtin-gateways.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["text"])])):T("",!0)]),default:e(()=>[c(),a(w,{data:[o],errors:[g]},{default:e(()=>{var r;return[a(p,{selected:(r=s.child())==null?void 0:r.name},C({_:2},[k(s.children,({name:n})=>({name:`${n}-tab`,fn:e(()=>[a(u,{to:{name:n}},{default:e(()=>[c(A(m(`builtin-gateways.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),a(d,null,{default:e(({Component:n})=>[(i(),l(B(n),{gateway:o},null,8,["gateway"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{E as default};
