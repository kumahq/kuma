import{d as V,e as t,o as c,m as l,w as e,a,l as D,as as T,c as C,a1 as R,p as S,b as m,T as k,J as A,t as X,E as y}from"./index-COT-_p62.js";const B={key:0},g=V({__name:"MeshExternalServiceDetailTabsView",setup(E){return(L,N)=>{const _=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),h=t("RouterView"),u=t("DataLoader"),v=t("AppView"),f=t("DataSource"),x=t("RouteView");return c(),l(x,{name:"mesh-external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:e(({route:s,t:n,uri:w})=>[a(f,{src:w(D(T),"/meshes/:mesh/mesh-external-service/:name",{mesh:s.params.mesh,name:s.params.service})},{default:e(({data:r,error:b})=>[a(v,{docs:n("services.mesh-external-service.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"mesh-external-service-list-view",params:{mesh:s.params.mesh}},text:n("services.routes.mesh-external-service-list-view.title")}]},{title:e(()=>[r?(c(),C("h1",B,[a(R,{text:s.params.service},{default:e(()=>[a(_,{title:n("services.routes.item.title",{name:r.name})},null,8,["title"])]),_:2},1032,["text"])])):S("",!0)]),default:e(()=>[m(),a(u,{data:[r],errors:[b]},{default:e(()=>{var i;return[a(d,{selected:(i=s.child())==null?void 0:i.name},k({_:2},[A(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[a(p,{to:{name:o}},{default:e(()=>[m(X(n(`services.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m(),a(h,null,{default:e(o=>[(c(),l(y(o.Component),{data:r},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{g as default};
