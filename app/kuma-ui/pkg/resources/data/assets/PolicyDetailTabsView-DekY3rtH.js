import{d as x,e as a,o as m,m as p,w as t,a as o,l as D,ao as v,c as C,$ as R,p as T,b as i,Q as P,J as k,t as A,E as S}from"./index-CgC5RQPZ.js";const X={key:0},g=x({__name:"PolicyDetailTabsView",setup($){return(B,L)=>{const r=a("RouteTitle"),_=a("XAction"),d=a("XTabs"),h=a("RouterView"),u=a("DataLoader"),f=a("AppView"),y=a("DataSource"),b=a("RouteView");return m(),p(b,{name:"policy-detail-tabs-view",params:{mesh:"",policy:"",policyPath:""}},{default:t(({route:e,t:c,uri:w})=>[o(y,{src:w(D(v),"/meshes/:mesh/policy-path/:path/policy/:name",{mesh:e.params.mesh,path:e.params.policyPath,name:e.params.policy})},{default:t(({data:s,error:V})=>[o(f,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"policy-list-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath}},text:c("policies.routes.item.breadcrumbs")}]},{title:t(()=>[s?(m(),C("h1",X,[o(R,{text:s.name},{default:t(()=>[o(r,{title:c("policies.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["text"])])):T("",!0)]),default:t(()=>[i(),o(u,{data:[s],errors:[V]},{default:t(()=>{var l;return[o(d,{selected:(l=e.child())==null?void 0:l.name},P({_:2},[k(e.children,({name:n})=>({name:`${n}-tab`,fn:t(()=>[o(_,{to:{name:n}},{default:t(()=>[i(A(c(`policies.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),o(h,null,{default:t(n=>[(m(),p(S(n.Component),{data:s},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{g as default};
