import{d as V,a as t,o as m,b as l,w as e,e as n,R as _,f as i,G as x,t as T,E as D,m as R,T as C}from"./index-BRR4OZXP.js";const S=V({__name:"ZoneIngressDetailTabsView",setup(h){return(y,A)=>{const p=t("RouteTitle"),u=t("XAction"),d=t("XTabs"),b=t("RouterView"),f=t("DataLoader"),z=t("AppView"),w=t("DataSource"),g=t("RouteView");return m(),l(g,{name:"zone-ingress-detail-tabs-view",params:{zone:"",zoneIngress:""}},{default:e(({route:a,t:r})=>[n(w,{src:`/zone-ingress-overviews/${a.params.zoneIngress}`},{default:e(({data:o,error:v})=>[n(z,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:r("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-ingress-list-view",params:{zone:a.params.zone}},text:r("zone-ingresses.routes.item.breadcrumbs")}]},_({default:e(()=>[i(),n(f,{data:[o],errors:[v]},{default:e(()=>{var c;return[n(d,{selected:(c=a.active)==null?void 0:c.name,"data-testid":"zone-ingress-tabs"},_({_:2},[x(a.children,({name:s})=>({name:`${s}-tab`,fn:e(()=>[n(u,{to:{name:s}},{default:e(()=>[i(T(r(`zone-ingresses.routes.item.navigation.${s}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),i(),n(b,null,{default:e(s=>[(m(),l(D(s.Component),{data:o},null,8,["data"]))]),_:2},1024)]}),_:2},1032,["data","errors"])]),_:2},[o?{name:"title",fn:e(()=>[R("h1",null,[n(C,{text:o.name},{default:e(()=>[n(p,{title:r("zone-ingresses.routes.item.title",{name:o.name})},null,8,["title"])]),_:2},1032,["text"])])]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:2},1032,["src"])]),_:1})}}});export{S as default};
