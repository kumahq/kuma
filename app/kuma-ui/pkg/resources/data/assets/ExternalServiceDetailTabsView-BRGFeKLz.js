import{d as u,e as t,o as v,m as x,w as s,a,k as h,a1 as b,b as o,R as w,J as f,t as V}from"./index-bM6gVJZj.js";const k=u({__name:"ExternalServiceDetailTabsView",setup(R){return(T,A)=>{const i=t("RouteTitle"),m=t("XAction"),l=t("XTabs"),p=t("RouterView"),_=t("AppView"),d=t("RouteView");return v(),x(d,{name:"external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:s(({route:e,t:n})=>[a(_,{docs:n("external-services.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"external-service-list-view",params:{mesh:e.params.mesh}},text:n("external-services.routes.item.breadcrumbs")}]},{title:s(()=>[h("h1",null,[a(b,{text:e.params.service},{default:s(()=>[a(i,{title:n("external-services.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:s(()=>{var c;return[o(),a(l,{selected:(c=e.child())==null?void 0:c.name},w({_:2},[f(e.children,({name:r})=>({name:`${r}-tab`,fn:s(()=>[a(m,{to:{name:r}},{default:s(()=>[o(V(n(`external-services.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),o(),a(p)]}),_:2},1032,["docs","breadcrumbs"])]),_:1})}}});export{k as default};
