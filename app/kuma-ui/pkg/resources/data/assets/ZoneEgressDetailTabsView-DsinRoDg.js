import{d as v,r as t,o as c,m as l,w as e,b as o,c as x,R as h,p as D,e as i,Q as R,s as C,t as T,E}from"./index--1DEc0sn.js";const k={key:0},y=v({__name:"ZoneEgressDetailTabsView",setup(A){return(S,X)=>{const _=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),u=t("RouterView"),z=t("DataLoader"),b=t("AppView"),f=t("DataSource"),w=t("RouteView");return c(),l(w,{name:"zone-egress-detail-tabs-view",params:{zone:"",zoneEgress:""}},{default:e(({route:s,can:g,t:n})=>[o(f,{src:`/zone-egress-overviews/${s.params.zoneEgress}`},{default:e(({data:a,error:V})=>[o(b,{docs:n("zone-ingresses.href.docs"),breadcrumbs:[...g("use zones")?[{to:{name:"zone-cp-list-view"},text:n("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:s.params.zone}},text:s.params.zone}]:[],{to:{name:"zone-egress-list-view",params:{zone:s.params.zone}},text:n("zone-egresses.routes.item.breadcrumbs")}]},{title:e(()=>[a?(c(),x("h1",k,[o(h,{text:a.name},{default:e(()=>[o(_,{title:n("zone-egresses.routes.item.title",{name:a.name})},null,8,["title"])]),_:2},1032,["text"])])):D("",!0)]),default:e(()=>[i(),o(z,{data:[a],errors:[V]},{default:e(()=>{var m;return[o(d,{selected:(m=s.child())==null?void 0:m.name},R({_:2},[C(s.children,({name:r})=>({name:`${r}-tab`,fn:e(()=>[o(p,{to:{name:r}},{default:e(()=>[i(T(n(`zone-egresses.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),o(u,null,{default:e(r=>[(c(),l(E(r.Component),{data:a},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{y as default};
