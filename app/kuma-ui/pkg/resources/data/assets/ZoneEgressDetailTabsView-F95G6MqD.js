import{d as x,h as t,o as m,a as l,w as e,j as o,V as _,k as c,B as D,t as h,r as R,g as T,X as C}from"./index-9gITI0JG.js";const S=x({__name:"ZoneEgressDetailTabsView",setup(X){return(k,A)=>{const p=t("RouteTitle"),u=t("XAction"),d=t("XTabs"),z=t("RouterView"),f=t("DataLoader"),w=t("AppView"),b=t("DataSource"),g=t("RouteView");return m(),l(g,{name:"zone-egress-detail-tabs-view",params:{zone:"",zoneEgress:""}},{default:e(({route:s,can:V,t:r})=>[o(b,{src:`/zone-egress-overviews/${s.params.zoneEgress}`},{default:e(({data:n,error:v})=>[o(w,{breadcrumbs:[...V("use zones")?[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-cp-detail-view",params:{zone:s.params.zone}},text:s.params.zone}]:[],{to:{name:"zone-egress-list-view",params:{zone:s.params.zone}},text:r("zone-egresses.routes.item.breadcrumbs")}]},_({default:e(()=>[c(),o(f,{data:[n],errors:[v]},{default:e(()=>{var i;return[o(d,{selected:(i=s.child())==null?void 0:i.name},_({_:2},[D(s.children,({name:a})=>({name:`${a}-tab`,fn:e(()=>[o(u,{to:{name:a}},{default:e(()=>[c(h(r(`zone-egresses.routes.item.navigation.${a}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),o(z,null,{default:e(a=>[(m(),l(R(a.Component),{data:n},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[n?{name:"title",fn:e(()=>[T("h1",null,[o(C,{text:n.name},{default:e(()=>[o(p,{title:r("zone-egresses.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{S as default};
