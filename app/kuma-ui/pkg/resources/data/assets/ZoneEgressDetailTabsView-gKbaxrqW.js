import{d as x,a as t,o as m,b as l,w as e,e as o,N as _,f as c,C as T,t as C,B as D,m as R,T as h}from"./index-cbxC5pCb.js";const S=x({__name:"ZoneEgressDetailTabsView",setup(B){return(y,A)=>{const u=t("RouteTitle"),p=t("XAction"),d=t("XTabs"),b=t("RouterView"),f=t("DataLoader"),z=t("AppView"),w=t("DataSource"),g=t("RouteView");return m(),l(g,{name:"zone-egress-detail-tabs-view",params:{zone:"",zoneEgress:""}},{default:e(({route:a,can:v,t:r})=>[o(w,{src:`/zone-egress-overviews/${a.params.zoneEgress}`},{default:e(({data:s,error:V})=>[o(z,{breadcrumbs:[...v("use zones")?[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")}]:[],{to:{name:"zone-egress-list-view",params:{zone:a.params.zone}},text:r("zone-egresses.routes.item.breadcrumbs")}]},_({default:e(()=>[c(),o(f,{data:[s],errors:[V]},{default:e(()=>{var i;return[o(d,{selected:(i=a.active)==null?void 0:i.name},_({_:2},[T(a.children,({name:n})=>({name:`${n}-tab`,fn:e(()=>[o(p,{to:{name:n}},{default:e(()=>[c(C(r(`zone-egresses.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),o(b,null,{default:e(n=>[(m(),l(D(n.Component),{data:s},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[s?{name:"title",fn:e(()=>[R("h1",null,[o(h,{text:s.name},{default:e(()=>[o(u,{title:r("zone-egresses.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{S as default};
