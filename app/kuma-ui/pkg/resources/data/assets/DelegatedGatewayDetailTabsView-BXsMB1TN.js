import{d as w,r as t,o as u,m as g,w as a,b as s,k as h,T as b,e as c,R as f,s as V,t as v}from"./index-DxrN05KS.js";const A=w({__name:"DelegatedGatewayDetailTabsView",setup(x){return(y,R)=>{const m=t("RouteTitle"),r=t("XAction"),d=t("XTabs"),l=t("RouterView"),p=t("AppView"),_=t("RouteView");return u(),g(_,{name:"delegated-gateway-detail-tabs-view",params:{mesh:"",service:""}},{default:a(({route:e,t:o})=>[s(p,{docs:o("delegated-gateways.href.docs"),breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"delegated-gateway-list-view",params:{mesh:e.params.mesh}},text:o("delegated-gateways.routes.item.breadcrumbs")}]},{title:a(()=>[h("h1",null,[s(b,{text:e.params.service},{default:a(()=>[s(m,{title:o("delegated-gateways.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:a(()=>{var i;return[c(),s(d,{selected:(i=e.child())==null?void 0:i.name},f({_:2},[V(e.children,({name:n})=>({name:`${n}-tab`,fn:a(()=>[s(r,{to:{name:n}},{default:a(()=>[c(v(o(`delegated-gateways.routes.item.navigation.${n}`)),1)]),_:2},1032,["to"])])}))]),1032,["selected"]),c(),s(l)]}),_:2},1032,["docs","breadcrumbs"])]),_:1})}}});export{A as default};
