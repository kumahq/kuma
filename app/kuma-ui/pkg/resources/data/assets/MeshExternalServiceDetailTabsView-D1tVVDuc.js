import{d as D,e as t,o as c,m as p,w as e,a,l as R,aq as C,c as T,a0 as S,p as k,b as i,R as A,K as X,t as y,F as B}from"./index-CKcsX_-l.js";const L={key:0},E=D({__name:"MeshExternalServiceDetailTabsView",setup(N){return($,n)=>{const _=t("RouteTitle"),d=t("XAction"),u=t("XTabs"),h=t("RouterView"),v=t("DataLoader"),f=t("AppView"),x=t("DataSource"),w=t("RouteView");return c(),p(w,{name:"mesh-external-service-detail-tabs-view",params:{mesh:"",service:""}},{default:e(({route:s,t:r,uri:b})=>[a(x,{src:b(R(C),"/meshes/:mesh/mesh-external-service/:name",{mesh:s.params.mesh,name:s.params.service})},{default:e(({data:m,error:V})=>[a(f,{docs:r("services.mesh-external-service.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"mesh-external-service-list-view",params:{mesh:s.params.mesh}},text:r("services.routes.mesh-external-service-list-view.title")}]},{title:e(()=>[m?(c(),T("h1",L,[a(S,{text:s.params.service},{default:e(()=>[a(_,{title:r("services.routes.item.title",{name:m.name})},null,8,["title"])]),_:2},1032,["text"])])):k("",!0)]),default:e(()=>[n[1]||(n[1]=i()),a(v,{data:[m],errors:[V]},{default:e(()=>{var l;return[a(u,{selected:(l=s.child())==null?void 0:l.name},A({_:2},[X(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[a(d,{to:{name:o}},{default:e(()=>[i(y(r(`services.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),n[0]||(n[0]=i()),a(h,null,{default:e(o=>[(c(),p(B(o.Component),{data:m},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{E as default};
