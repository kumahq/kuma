import{d as p,r as _,o as a,a as i,w as t,h as o,b as s,C as f,s as v,g as n,f as z,i as g,E as w,t as N}from"./index-2a9ba339.js";import{E as b,g as x,e as V,A as C,_ as h}from"./RouteView.vue_vue_type_script_setup_true_lang-5d6806ed.js";import{_ as k}from"./RouteTitle.vue_vue_type_script_setup_true_lang-4859e7c4.js";import{N as y}from"./NavTabs-e2692c17.js";const R=p({__name:"ZoneIndexView",setup(E){const r=b(),{t:e}=x(),l=V(),u=[{title:e("zones.routes.items.navigation.zone-cp-list-view"),routeName:"zone-cp-list-view",module:"zone-cps"},{title:e("zones.routes.items.navigation.zone-ingress-list-view"),routeName:"zone-ingress-list-view",module:"zone-ingresses"},{title:e("zones.routes.items.navigation.zone-egress-list-view"),routeName:"zone-egress-list-view",module:"zone-egresses"}];return(A,S)=>{const c=_("RouterView");return a(),i(h,null,{default:t(()=>[o(C,{breadcrumbs:[{to:{name:"zone-index-view"},text:s(e)("zones.routes.items.breadcrumbs")}]},f({title:t(()=>[v("h1",null,[o(k,{title:s(e)("zones.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[n(),n(),s(l).getters["config/getMulticlusterStatus"]?(a(),i(y,{key:0,tabs:u})):z("",!0),n(),o(c,null,{default:t(({Component:m,route:d})=>[(a(),i(g(m),{key:d.path}))]),_:1})]),_:2},[s(r)("KUMA_ZONE_CREATION_FLOW")==="enabled"?{name:"actions",fn:t(()=>[o(s(w),{appearance:"creation",icon:"plus",to:{name:"zone-create-view"}},{default:t(()=>[n(N(s(e)("zones.index.create")),1)]),_:1})]),key:"0"}:void 0]),1032,["breadcrumbs"])]),_:1})}}});export{R as default};
