import{d as v,r as e,o as r,m as u,w as t,b as a,p as C,ao as D,c as T,q as B,e as c,T as R,v as X,t as k,K as A}from"./index-C-Llvxgw.js";const S={key:0},q=v({__name:"BuiltinGatewayDetailTabsView",setup(L){return(N,m)=>{const p=e("RouteTitle"),_=e("XCopyButton"),d=e("XAction"),w=e("XTabs"),b=e("RouterView"),y=e("DataLoader"),h=e("AppView"),f=e("DataSource"),g=e("RouteView");return r(),u(g,{name:"builtin-gateway-detail-tabs-view",params:{mesh:"",gateway:""}},{default:t(({route:s,t:i,uri:V})=>[a(f,{src:V(C(D),"/meshes/:mesh/mesh-gateways/:name",{mesh:s.params.mesh,name:s.params.gateway})},{default:t(({data:o,error:x})=>[a(h,{docs:i("builtin-gateways.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"builtin-gateway-list-view",params:{mesh:s.params.mesh}},text:i("builtin-gateways.routes.item.breadcrumbs")}]},{title:t(()=>[o?(r(),T("h1",S,[a(_,{text:o.name},{default:t(()=>[a(p,{title:i("builtin-gateways.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["text"])])):B("",!0)]),default:t(()=>[m[1]||(m[1]=c()),a(y,{data:[o],errors:[x]},{default:t(()=>{var l;return[a(w,{selected:(l=s.child())==null?void 0:l.name},R({_:2},[X(s.children,({name:n})=>({name:`${n}-tab`,fn:t(()=>[a(d,{to:{name:n}},{default:t(()=>[c(k(i(`builtin-gateways.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),m[0]||(m[0]=c()),a(b,null,{default:t(({Component:n})=>[(r(),u(A(n),{gateway:o},null,8,["gateway"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{q as default};
