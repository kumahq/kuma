import{d as D,r as t,o as i,p as l,w as o,b as a,m as R,ao as v,c as T,q as X,e as p,R as B,K as P,t as k,F as A}from"./index-BIN9nSPF.js";const S={key:0},$=D({__name:"PolicyDetailTabsView",setup(L){return(N,m)=>{const _=t("RouteTitle"),u=t("XCopyButton"),d=t("XAction"),h=t("XTabs"),y=t("RouterView"),f=t("DataLoader"),b=t("AppView"),w=t("DataSource"),V=t("RouteView");return i(),l(V,{name:"policy-detail-tabs-view",params:{mesh:"",policy:"",policyPath:""}},{default:o(({route:e,t:c,uri:x})=>[a(w,{src:x(R(v),"/meshes/:mesh/policy-path/:path/policy/:name",{mesh:e.params.mesh,path:e.params.policyPath,name:e.params.policy})},{default:o(({data:s,error:C})=>[a(b,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"policy-list-view",params:{mesh:e.params.mesh,policyPath:e.params.policyPath}},text:c("policies.routes.item.breadcrumbs")}]},{title:o(()=>[s?(i(),T("h1",S,[a(u,{text:s.name},{default:o(()=>[a(_,{title:c("policies.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["text"])])):X("",!0)]),default:o(()=>[m[1]||(m[1]=p()),a(f,{data:[s],errors:[C]},{default:o(()=>{var r;return[a(h,{selected:(r=e.child())==null?void 0:r.name},B({_:2},[P(e.children,({name:n})=>({name:`${n}-tab`,fn:o(()=>[a(d,{to:{name:n}},{default:o(()=>[p(k(c(`policies.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m[0]||(m[0]=p()),a(y,null,{default:o(n=>[(i(),l(A(n.Component),{data:s},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{$ as default};
