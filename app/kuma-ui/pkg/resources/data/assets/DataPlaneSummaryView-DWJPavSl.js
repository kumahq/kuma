import{d as N,e as t,o as m,p as c,w as e,a,l as d,t as u,b as r,c as b,J as g,K as w,n as k,R as v,F as A,r as B}from"./index-DH1Ug2X_.js";const E=N({__name:"DataPlaneSummaryView",props:{items:{},routeName:{}},setup(V){const _=V;return(P,n)=>{const C=t("XEmptyState"),D=t("RouteTitle"),y=t("XAction"),R=t("XTabs"),S=t("RouterView"),T=t("AppView"),X=t("DataCollection"),x=t("RouteView");return m(),c(x,{name:_.routeName,params:{dataPlane:""}},{default:e(({route:i,t:s})=>[a(X,{items:_.items,predicate:p=>p.id===i.params.dataPlane},{empty:e(()=>[a(C,null,{title:e(()=>[d("h2",null,u(s("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:e(()=>[n[0]||(n[0]=r()),d("p",null,u(s("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:e(({items:p})=>[(m(!0),b(g,null,w([p[0]],o=>(m(),c(T,{key:o.id},{title:e(()=>[d("h2",{class:k(`type-${o.dataplaneType}`)},[a(y,{to:{name:"data-plane-detail-view",params:{dataPlane:o.id}}},{default:e(()=>[a(D,{title:s("data-planes.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:e(()=>{var f;return[n[1]||(n[1]=r()),a(R,{selected:(f=i.child())==null?void 0:f.name},v({_:2},[w(i.children,({name:l})=>({name:`${l}-tab`,fn:e(()=>[a(y,{to:{name:l}},{default:e(()=>[r(u(s(`data-planes.routes.item.navigation.${l}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),n[2]||(n[2]=r()),a(S,null,{default:e(({Component:l})=>[(m(),c(A(l),{data:o},{default:e(()=>[B(P.$slots,"default")]),_:2},1032,["data"]))]),_:2},1024)]}),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:3},8,["name"])}}});export{E as default};
