import{d as v,e as t,o as c,k as l,w as e,a as o,c as x,$ as h,l as C,b as i,Q as D,G as R,t as T,C as k}from"./index-loxRIpgb.js";const $={key:0},y=v({__name:"ZoneEgressDetailTabsView",setup(A){return(E,S)=>{const _=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),u=t("RouterView"),z=t("DataLoader"),b=t("AppView"),f=t("DataSource"),w=t("RouteView");return c(),l(w,{name:"zone-egress-detail-tabs-view",params:{zone:"",zoneEgress:""}},{default:e(({route:s,can:g,t:n})=>[o(f,{src:`/zone-egress-overviews/${s.params.zoneEgress}`},{default:e(({data:a,error:V})=>[o(b,{docs:n("zone-ingresses.href.docs"),breadcrumbs:[...g("use zones")?[{to:{name:"zone-cp-list-view"},text:n("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:s.params.zone}},text:s.params.zone}]:[],{to:{name:"zone-egress-list-view",params:{zone:s.params.zone}},text:n("zone-egresses.routes.item.breadcrumbs")}]},{title:e(()=>[a?(c(),x("h1",$,[o(h,{text:a.name},{default:e(()=>[o(_,{title:n("zone-egresses.routes.item.title",{name:a.name})},null,8,["title"])]),_:2},1032,["text"])])):C("",!0)]),default:e(()=>[i(),o(z,{data:[a],errors:[V]},{default:e(()=>{var m;return[o(d,{selected:(m=s.child())==null?void 0:m.name},D({_:2},[R(s.children,({name:r})=>({name:`${r}-tab`,fn:e(()=>[o(p,{to:{name:r}},{default:e(()=>[i(T(n(`zone-egresses.routes.item.navigation.${r}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),o(u,null,{default:e(r=>[(c(),l(k(r.Component),{data:a},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},1032,["docs","breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{y as default};
