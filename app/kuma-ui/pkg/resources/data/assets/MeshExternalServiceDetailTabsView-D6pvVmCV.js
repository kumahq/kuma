import{d as V,e as t,o as c,k as l,w as e,a,j as C,aq as D,c as R,$ as T,l as k,b as m,Q as S,G as A,t as X,C as $}from"./index-loxRIpgb.js";const y={key:0},E=V({__name:"MeshExternalServiceDetailTabsView",setup(B){return(L,N)=>{const _=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),h=t("RouterView"),u=t("DataLoader"),v=t("AppView"),f=t("DataSource"),x=t("RouteView");return c(),l(x,{name:"mesh-external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:e(({route:s,t:n,uri:w})=>[a(f,{src:w(C(D),"/meshes/:mesh/mesh-external-service/:name",{mesh:s.params.mesh,name:s.params.service})},{default:e(({data:r,error:b})=>[a(v,{docs:n("services.mesh-external-service.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"mesh-external-service-list-view",params:{mesh:s.params.mesh}},text:n("services.routes.mesh-external-service-list-view.title")}]},{title:e(()=>[r?(c(),R("h1",y,[a(T,{text:s.params.service},{default:e(()=>[a(_,{title:n("services.routes.item.title",{name:r.name})},null,8,["title"])]),_:2},1032,["text"])])):k("",!0)]),default:e(()=>[m(),a(u,{data:[r],errors:[b]},{default:e(()=>{var i;return[a(d,{selected:(i=s.child())==null?void 0:i.name},S({_:2},[A(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[a(p,{to:{name:o}},{default:e(()=>[m(X(n(`services.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m(),a(h,null,{default:e(o=>[(c(),l($(o.Component),{data:r},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{E as default};
