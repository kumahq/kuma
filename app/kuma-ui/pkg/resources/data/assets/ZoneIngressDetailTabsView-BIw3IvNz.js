import{d as v,r as e,o as c,q as l,w as t,b as n,c as C,s as k,e as m,T as y,N as D,t as T,I as R}from"./index-BP47cGGe.js";const X={key:0},S=v({__name:"ZoneIngressDetailTabsView",setup(B){return(A,i)=>{const _=e("RouteTitle"),u=e("XCopyButton"),d=e("XAction"),z=e("XTabs"),w=e("RouterView"),b=e("DataLoader"),g=e("AppView"),f=e("DataSource"),x=e("RouteView");return c(),l(x,{name:"zone-ingress-detail-tabs-view",params:{zone:"",proxy:""}},{default:t(({route:s,t:r})=>[n(f,{src:`/zone-ingress-overviews/${s.params.proxy}`},{default:t(({data:o,error:V})=>[n(g,{docs:r("zone-ingresses.href.docs"),breadcrumbs:[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:s.params.zone}},text:s.params.zone},{to:{name:"zone-ingress-list-view",params:{zone:s.params.zone}},text:r("zone-ingresses.routes.item.breadcrumbs")}]},{title:t(()=>[o?(c(),C("h1",X,[n(u,{text:o.name},{default:t(()=>[n(_,{title:r("zone-ingresses.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["text"])])):k("",!0)]),default:t(()=>[i[1]||(i[1]=m()),n(b,{data:[o],errors:[V]},{default:t(()=>{var p;return[n(z,{selected:(p=s.child())==null?void 0:p.name,"data-testid":"zone-ingress-tabs"},y({_:2},[D(s.children,({name:a})=>({name:`${a}-tab`,fn:t(()=>[n(d,{to:{name:a}},{default:t(()=>[m(T(r(`zone-ingresses.routes.item.navigation.${a}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i[0]||(i[0]=m()),n(w,null,{default:t(a=>[(c(),l(R(a.Component),{networking:o==null?void 0:o.zoneIngress.networking,data:o},null,8,["networking","data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{S as default};
