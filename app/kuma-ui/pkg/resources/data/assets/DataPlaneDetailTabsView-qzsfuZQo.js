import{d as T,e as a,o as c,m as i,w as e,a as t,T as p,b as r,J as R,t as C,E as $,k,a1 as A}from"./index-COT-_p62.js";const B=T({__name:"DataPlaneDetailTabsView",props:{mesh:{}},setup(d){const _=d;return(S,X)=>{const u=a("RouteTitle"),h=a("XAction"),f=a("XTabs"),w=a("RouterView"),b=a("DataLoader"),V=a("AppView"),v=a("DataSource"),x=a("RouteView");return c(),i(x,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:e(({route:s,t:m})=>[t(v,{src:`/meshes/${s.params.mesh}/dataplane-overviews/${s.params.dataPlane}`},{default:e(({data:n,error:D})=>[t(V,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"data-plane-list-view",params:{mesh:s.params.mesh}},text:m("data-planes.routes.item.breadcrumbs")}]},p({default:e(()=>[r(),t(b,{data:[n],errors:[D]},{default:e(()=>{var l;return[t(f,{selected:(l=s.child())==null?void 0:l.name},p({_:2},[R(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[t(h,{to:{name:o}},{default:e(()=>[r(C(m(`data-planes.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),r(),t(w,null,{default:e(o=>[(c(),i($(o.Component),{data:n,mesh:_.mesh},null,8,["data","mesh"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[n?{name:"title",fn:e(()=>[k("h1",null,[t(A,{text:n.name},{default:e(()=>[t(u,{title:m("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{B as default};
