import{d as x,e as a,o as i,k as r,w as t,a as o,j as C,ao as D,c as v,$ as R,l as T,b as m,Q as k,G as P,t as A,C as S}from"./index-CUmbT3FY.js";const X={key:0},g=x({__name:"PolicyDetailTabsView",setup($){return(B,L)=>{const p=a("RouteTitle"),_=a("XAction"),d=a("XTabs"),h=a("RouterView"),u=a("DataLoader"),f=a("AppView"),y=a("DataSource"),b=a("RouteView");return i(),r(b,{name:"policy-detail-tabs-view",params:{mesh:"",policy:"",policyPath:""}},{default:t(({route:e,t:c,uri:w})=>[o(y,{src:w(C(D),"/meshes/:mesh/policy-path/:path/policy/:name",{mesh:e.params.mesh,path:e.params.policyPath,name:e.params.policy})},{default:t(({data:s,error:V})=>[o(f,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"policy-list-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath}},text:c("policies.routes.item.breadcrumbs")}]},{title:t(()=>[s?(i(),v("h1",X,[o(R,{text:s.name},{default:t(()=>[o(p,{title:c("policies.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["text"])])):T("",!0)]),default:t(()=>[m(),o(u,{data:[s],errors:[V]},{default:t(()=>{var l;return[o(d,{selected:(l=e.child())==null?void 0:l.name},k({_:2},[P(e.children,({name:n})=>({name:`${n}-tab`,fn:t(()=>[o(_,{to:{name:n}},{default:t(()=>[m(A(c(`policies.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m(),o(h,null,{default:t(n=>[(i(),r(S(n.Component),{data:s},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{g as default};
