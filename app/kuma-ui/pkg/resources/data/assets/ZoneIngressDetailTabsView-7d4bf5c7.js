import{d as V,k as x,U as R,a as s,o as n,b as c,w as t,e as o,l,m as k,f as w,c as C,F as T,C as h}from"./index-0b4678e0.js";import{E as y}from"./ErrorBlock-1570e211.js";import{_ as B}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-6507e090.js";import{N as I}from"./NavTabs-8a79df3a.js";import{T as N}from"./TextWithCopyButton-bc8c6ef3.js";import"./index-fce48c05.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-0acf695d.js";import"./CopyButton-18f43ddc.js";const G=V({__name:"ZoneIngressDetailTabsView",setup(D){var _;const{t:a}=x(),z=(((_=R().getRoutes().find(e=>e.name==="zone-ingress-detail-tabs-view"))==null?void 0:_.children)??[]).map(e=>{var i,m;const u=typeof e.name>"u"?(i=e.children)==null?void 0:i[0]:e,r=u.name,p=((m=u.meta)==null?void 0:m.module)??"";return{title:a(`zone-ingresses.routes.item.navigation.${r}`),routeName:r,module:p}});return(e,u)=>{const r=s("RouteTitle"),p=s("RouterView"),f=s("DataSource"),i=s("AppView"),m=s("RouteView");return n(),c(m,{name:"zone-ingress-detail-tabs-view",params:{zoneIngress:""}},{default:t(({route:d})=>[o(i,{breadcrumbs:[{to:{name:"zone-cp-list-view"},text:l(a)("zone-cps.routes.item.breadcrumbs")},{to:{name:"zone-ingress-list-view"},text:l(a)("zone-ingresses.routes.item.breadcrumbs")}]},{title:t(()=>[k("h1",null,[o(N,{text:d.params.zoneIngress},{default:t(()=>[o(r,{title:l(a)("zone-ingresses.routes.item.title",{name:d.params.zoneIngress})},null,8,["title"])]),_:2},1032,["text"])])]),default:t(()=>[w(),o(f,{src:`/zone-ingress-overviews/${d.params.zoneIngress}`},{default:t(({data:b,error:g})=>[g!==void 0?(n(),c(y,{key:0,error:g},null,8,["error"])):b===void 0?(n(),c(B,{key:1})):(n(),C(T,{key:2},[o(I,{class:"route-zone-ingress-detail-view-tabs",tabs:l(z)},null,8,["tabs"]),w(),o(p,null,{default:t(v=>[(n(),c(h(v.Component),{data:b},null,8,["data"]))]),_:2},1024)],64))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1})}}});export{G as default};
