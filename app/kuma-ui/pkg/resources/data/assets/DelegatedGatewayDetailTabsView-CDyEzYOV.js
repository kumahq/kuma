import{d as u,e as t,o as g,m as b,w as a,a as s,k as f,a0 as V,b as m,R as h,K as v,t as x}from"./index-DpJ_igul.js";const A=u({__name:"DelegatedGatewayDetailTabsView",setup(y){return(R,o)=>{const d=t("RouteTitle"),l=t("XAction"),c=t("XTabs"),p=t("RouterView"),_=t("AppView"),w=t("RouteView");return g(),b(w,{name:"delegated-gateway-detail-tabs-view",params:{mesh:"",service:""}},{default:a(({route:e,t:n})=>[s(_,{docs:n("delegated-gateways.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"delegated-gateway-list-view",params:{mesh:e.params.mesh}},text:n("delegated-gateways.routes.item.breadcrumbs")}]},{title:a(()=>[f("h1",null,[s(V,{text:e.params.service},{default:a(()=>[s(d,{title:n("delegated-gateways.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:a(()=>{var r;return[o[0]||(o[0]=m()),s(c,{selected:(r=e.child())==null?void 0:r.name},h({_:2},[v(e.children,({name:i})=>({name:`${i}-tab`,fn:a(()=>[s(l,{to:{name:i}},{default:a(()=>[m(x(n(`delegated-gateways.routes.item.navigation.${i}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),o[1]||(o[1]=m()),s(p)]}),_:2},1032,["docs","breadcrumbs"])]),_:1})}}});export{A as default};
