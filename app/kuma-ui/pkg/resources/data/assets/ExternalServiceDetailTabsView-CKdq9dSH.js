import{d as u,h as t,o as v,a as x,w as a,j as s,g as h,a3 as w,k as o,P as b,B as f,t as V}from"./index-CMlVV7ds.js";const X=u({__name:"ExternalServiceDetailTabsView",setup(R){return(T,A)=>{const c=t("RouteTitle"),m=t("XAction"),l=t("XTabs"),p=t("RouterView"),_=t("AppView"),d=t("RouteView");return v(),x(d,{name:"external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:a(({route:e,t:n})=>[s(_,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"external-service-list-view",params:{mesh:e.params.mesh}},text:n("external-services.routes.item.breadcrumbs")}]},{title:a(()=>[h("h1",null,[s(w,{text:e.params.service},{default:a(()=>[s(c,{title:n("external-services.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:a(()=>{var i;return[o(),s(l,{selected:(i=e.child())==null?void 0:i.name},b({_:2},[f(e.children,({name:r})=>({name:`${r}-tab`,fn:a(()=>[s(m,{to:{name:r}},{default:a(()=>[o(V(n(`external-services.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),o(),s(p)]}),_:2},1032,["breadcrumbs"])]),_:1})}}});export{X as default};
