import{d as _,l as f,R as g,a as n,o as b,b as h,w as m,q as c,e as a,m as v,f as p}from"./index-CqXGLTiP.js";import{N as R}from"./NavTabs-M3eNsnqf.js";import{T as V}from"./TextWithCopyButton-3lrZWnHN.js";import"./CopyButton-MfTze-YE.js";import"./index-FZCiQto1.js";const D=_({__name:"DelegatedGatewayDetailTabsView",setup(x){var d;const{t:r}=f(),w=(((d=g().getRoutes().find(t=>t.name==="delegated-gateway-detail-tabs-view"))==null?void 0:d.children)??[]).map(t=>{var o,e;const i=typeof t.name>"u"?(o=t.children)==null?void 0:o[0]:t,s=i.name,l=((e=i.meta)==null?void 0:e.module)??"";return{title:r(`delegated-gateways.routes.item.navigation.${s}`),routeName:s,module:l}});return(t,i)=>{const s=n("RouteTitle"),l=n("RouterView"),u=n("AppView"),o=n("RouteView");return b(),h(o,{name:"delegated-gateway-detail-tabs-view",params:{mesh:"",service:""}},{default:m(({route:e})=>[a(u,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:e.params.mesh}},text:e.params.mesh},{to:{name:"delegated-gateway-list-view",params:{mesh:e.params.mesh}},text:c(r)("delegated-gateways.routes.item.breadcrumbs")}]},{title:m(()=>[v("h1",null,[a(V,{text:e.params.service},{default:m(()=>[a(s,{title:c(r)("delegated-gateways.routes.item.title",{name:e.params.service})},null,8,["title"])]),_:2},1032,["text"])])]),default:m(()=>[p(),a(R,{tabs:c(w)},null,8,["tabs"]),p(),a(l)]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{D as default};
