import{d as x,e as t,o as r,k as l,w as e,a,j as C,ap as D,c as R,$ as T,l as k,b as m,Q as S,G as A,t as X,C as $}from"./index-CUmbT3FY.js";const y={key:0},j=x({__name:"MeshServiceDetailTabsView",setup(B){return(L,N)=>{const _=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),h=t("RouterView"),u=t("DataLoader"),v=t("AppView"),f=t("DataSource"),w=t("RouteView");return r(),l(w,{name:"mesh-service-detail-tabs-view",params:{mesh:"",service:""}},{default:e(({route:s,t:n,uri:b})=>[a(f,{src:b(C(D),"/meshes/:mesh/mesh-service/:name",{mesh:s.params.mesh,name:s.params.service})},{default:e(({data:c,error:V})=>[a(v,{docs:n("services.mesh-service.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"mesh-service-list-view",params:{mesh:s.params.mesh}},text:n("services.routes.mesh-service-list-view.title")}]},{title:e(()=>[c?(r(),R("h1",y,[a(T,{text:s.params.service},{default:e(()=>[a(_,{title:n("services.routes.item.title",{name:c.name})},null,8,["title"])]),_:2},1032,["text"])])):k("",!0)]),default:e(()=>[m(),a(u,{data:[c],errors:[V]},{default:e(()=>{var i;return[a(d,{selected:(i=s.child())==null?void 0:i.name},S({_:2},[A(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[a(p,{to:{name:o}},{default:e(()=>[m(X(n(`services.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m(),a(h,null,{default:e(o=>[(r(),l($(o.Component),{data:c},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{j as default};
