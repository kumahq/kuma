import{_ as x,r as t,m as b,o as f,w as s,b as a,e as i,M as w,v as V,t as h,q as R}from"./index-D_WxlpfD.js";const T={};function X(B,o){const m=t("RouteTitle"),l=t("XCopyButton"),p=t("XAction"),_=t("XTabs"),d=t("RouterView"),u=t("AppView"),v=t("RouteView");return f(),b(v,{name:"external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:s(({route:e,t:n})=>[a(u,{docs:n("external-services.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"external-service-list-view",params:{mesh:e.params.mesh}},text:n("external-services.routes.item.breadcrumbs")}]},{title:s(()=>[R("h1",null,[a(l,{text:e.params.service},{default:s(()=>[a(m,{title:n("external-services.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:s(()=>{var c;return[o[0]||(o[0]=i()),a(_,{selected:(c=e.child())==null?void 0:c.name},w({_:2},[V(e.children,({name:r})=>({name:`${r}-tab`,fn:s(()=>[a(p,{to:{name:r}},{default:s(()=>[i(h(n(`external-services.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),o[1]||(o[1]=i()),a(d)]}),_:2},1032,["docs","breadcrumbs"])]),_:1})}const C=x(T,[["render",X]]);export{C as default};
