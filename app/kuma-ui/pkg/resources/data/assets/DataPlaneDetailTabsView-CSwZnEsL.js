import{d as x,i as a,o as c,a as i,w as e,j as t,P as p,k as r,J as D,t as R,r as T,g as C,Z as $}from"./index-B4OAi35c.js";const X=x({__name:"DataPlaneDetailTabsView",setup(k){return(A,P)=>{const _=a("RouteTitle"),d=a("XAction"),u=a("XTabs"),f=a("RouterView"),h=a("DataLoader"),w=a("AppView"),b=a("DataSource"),V=a("RouteView");return c(),i(V,{name:"data-plane-detail-tabs-view",params:{mesh:"",dataPlane:""}},{default:e(({route:s,t:m})=>[t(b,{src:`/meshes/${s.params.mesh}/dataplane-overviews/${s.params.dataPlane}`},{default:e(({data:n,error:v})=>[t(w,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"data-plane-list-view",params:{mesh:s.params.mesh}},text:m("data-planes.routes.item.breadcrumbs")}]},p({default:e(()=>[r(),t(h,{data:[n],errors:[v]},{default:e(()=>{var l;return[t(u,{selected:(l=s.child())==null?void 0:l.name},p({_:2},[D(s.children,({name:o})=>({name:`${o}-tab`,fn:e(()=>[t(d,{to:{name:o}},{default:e(()=>[r(R(m(`data-planes.routes.item.navigation.${o}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),r(),t(f,null,{default:e(o=>[(c(),i(T(o.Component),{data:n},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[n?{name:"title",fn:e(()=>[C("h1",null,[t($,{text:n.name},{default:e(()=>[t(_,{title:m("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{X as default};
