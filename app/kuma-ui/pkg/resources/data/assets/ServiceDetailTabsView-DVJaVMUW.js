import{d as v,r as t,o as b,p as w,w as s,b as a,l as f,V,e as r,R as h,K as x,t as R}from"./index-gI7YoWPY.js";const B=v({__name:"ServiceDetailTabsView",setup(T){return(A,o)=>{const m=t("RouteTitle"),l=t("XAction"),p=t("XTabs"),d=t("RouterView"),_=t("AppView"),u=t("RouteView");return b(),w(u,{name:"service-detail-tabs-view",params:{mesh:"",service:""}},{default:s(({route:e,t:i})=>[a(_,{docs:i("services.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"service-list-view",params:{mesh:e.params.mesh}},text:i("services.routes.item.breadcrumbs")}]},{title:s(()=>[f("h1",null,[a(V,{text:e.params.service},{default:s(()=>[a(m,{title:i("services.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:s(()=>{var c;return[o[0]||(o[0]=r()),a(p,{selected:(c=e.child())==null?void 0:c.name},h({_:2},[x(e.children,({name:n})=>({name:`${n}-tab`,fn:s(()=>[a(l,{to:{name:n}},{default:s(()=>[r(R(i(`services.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),o[1]||(o[1]=r()),a(d)]}),_:2},1032,["docs","breadcrumbs"])]),_:1})}}});export{B as default};
