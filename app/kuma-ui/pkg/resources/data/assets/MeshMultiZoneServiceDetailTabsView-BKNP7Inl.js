import{d as D,r as e,o as i,q as p,w as t,b as o,p as T,ap as R,c as X,s as y,e as c,T as z,N as B,t as S,I as k}from"./index-oTPgN0we.js";const A={key:0},M=D({__name:"MeshMultiZoneServiceDetailTabsView",setup(N){return(L,n)=>{const u=e("RouteTitle"),_=e("XCopyButton"),d=e("XAction"),h=e("XTabs"),v=e("RouterView"),f=e("DataLoader"),w=e("AppView"),b=e("DataSource"),V=e("RouteView");return i(),p(V,{name:"mesh-multi-zone-service-detail-tabs-view",params:{mesh:"",service:""}},{default:t(({route:s,t:r,uri:x})=>[o(b,{src:x(T(R),"/meshes/:mesh/mesh-multi-zone-service/:name",{mesh:s.params.mesh,name:s.params.service})},{default:t(({data:m,error:C})=>[o(w,{docs:r("services.mesh-multi-zone-service.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"mesh-multi-zone-service-list-view",params:{mesh:s.params.mesh}},text:r("services.routes.mesh-multi-zone-service-list-view.title")}]},{title:t(()=>[m?(i(),X("h1",A,[o(_,{text:s.params.service},{default:t(()=>[o(u,{title:r("services.routes.item.title",{name:m.name})},null,8,["title"])]),_:2},1032,["text"])])):y("",!0)]),default:t(()=>[n[1]||(n[1]=c()),o(f,{data:[m],errors:[C]},{default:t(()=>{var l;return[o(h,{selected:(l=s.child())==null?void 0:l.name},z({_:2},[B(s.children,({name:a})=>({name:`${a}-tab`,fn:t(()=>[o(d,{to:{name:a}},{default:t(()=>[c(S(r(`services.routes.item.navigation.${a}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),n[0]||(n[0]=c()),o(v,null,{default:t(a=>[(i(),p(k(a.Component),{data:m},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{M as default};
