import{d as v,e as t,o as c,m as p,w as e,a as o,c as x,$ as D,p as C,b as m,Q as R,J as T,t as $,E as k}from"./index-Dx_kP1mI.js";const A={key:0},y=v({__name:"ZoneIngressDetailTabsView",setup(S){return(X,i)=>{const _=t("RouteTitle"),d=t("XAction"),u=t("XTabs"),z=t("RouterView"),b=t("DataLoader"),f=t("AppView"),w=t("DataSource"),g=t("RouteView");return c(),p(g,{name:"zone-ingress-detail-tabs-view",params:{zone:"",zoneIngress:""}},{default:e(({route:s,t:n})=>[o(w,{src:`/zone-ingress-overviews/${s.params.zoneIngress}`},{default:e(({data:a,error:V})=>[o(f,{docs:n("zone-ingresses.href.docs"),breadcrumbs:[{to:{name:"zone-cp-list-view"},text:n("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:s.params.zone}},text:s.params.zone},{to:{name:"zone-ingress-list-view",params:{zone:s.params.zone}},text:n("zone-ingresses.routes.item.breadcrumbs")}]},{title:e(()=>[a?(c(),x("h1",A,[o(D,{text:a.name},{default:e(()=>[o(_,{title:n("zone-ingresses.routes.item.title",{name:a.name})},null,8,["title"])]),_:2},1032,["text"])])):C("",!0)]),default:e(()=>[i[1]||(i[1]=m()),o(b,{data:[a],errors:[V]},{default:e(()=>{var l;return[o(u,{selected:(l=s.child())==null?void 0:l.name,"data-testid":"zone-ingress-tabs"},R({_:2},[T(s.children,({name:r})=>({name:`${r}-tab`,fn:e(()=>[o(d,{to:{name:r}},{default:e(()=>[m($(n(`zone-ingresses.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i[0]||(i[0]=m()),o(z,null,{default:e(r=>[(c(),p(k(r.Component),{data:a},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{y as default};