import{d as x,a as t,o as m,b as l,w as e,e as o,O as _,f as c,G as D,t as R,E as T,m as h,a1 as C}from"./index-UmH8j8ci.js";const $=x({__name:"ZoneEgressDetailTabsView",setup(A){return(E,S)=>{const u=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),f=t("RouterView"),b=t("DataLoader"),z=t("AppView"),w=t("DataSource"),g=t("RouteView");return m(),l(g,{name:"zone-egress-detail-tabs-view",params:{zone:"",zoneEgress:""}},{default:e(({route:a,can:V,t:r})=>[o(w,{src:`/zone-egress-overviews/${a.params.zoneEgress}`},{default:e(({data:s,error:v})=>[o(z,{breadcrumbs:[...V("use zones")?[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")}]:[],{to:{name:"zone-egress-list-view",params:{zone:a.params.zone}},text:r("zone-egresses.routes.item.breadcrumbs")}]},_({default:e(()=>[c(),o(b,{data:[s],errors:[v]},{default:e(()=>{var i;return[o(d,{selected:(i=a.child())==null?void 0:i.name},_({_:2},[D(a.children,({name:n})=>({name:`${n}-tab`,fn:e(()=>[o(p,{to:{name:n}},{default:e(()=>[c(R(r(`zone-egresses.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),o(f,null,{default:e(n=>[(m(),l(T(n.Component),{data:s},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[s?{name:"title",fn:e(()=>[h("h1",null,[o(C,{text:s.name},{default:e(()=>[o(u,{title:r("zone-egresses.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{$ as default};
