import{d as g,u as k,q as n,o as e,a as t,w as m,h as i,b as c,g as w,k as z,e as b}from"./index-0b8ed13f.js";import{_ as y}from"./ZoneIngressDetails.vue_vue_type_script_setup_true_lang-8bbefd03.js";import{j as h,k as I,g as $,_ as x}from"./RouteView.vue_vue_type_script_setup_true_lang-4999f19d.js";import{_ as B}from"./RouteTitle.vue_vue_type_script_setup_true_lang-65d5caa7.js";import{_ as E}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-0272b2d5.js";import{E as V}from"./ErrorBlock-ddeaa4b4.js";import{_ as N}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-bcba5a7a.js";import"./DefinitionListItem-a18dd4eb.js";import"./EnvoyData-e6f53ad1.js";import"./kongponents.es-a99534bb.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-e187911d.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-eb706f57.js";import"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-80496b3d.js";import"./TabsWidget-8fc36bbe.js";import"./QueryParameter-70743f73.js";import"./TextWithCopyButton-45a80278.js";const A={class:"zone-details"},C={key:3,class:"kcard-border","data-testid":"detail-view-details"},U=g({__name:"ZoneIngressDetailView",setup(D){const _=h(),f=k(),{t:l}=I(),s=n(null),o=n(!0),r=n(null);p();function p(){d()}async function d(){o.value=!0,r.value=null;const u=f.params.zoneIngress;try{s.value=await _.getZoneIngressOverview({name:u})}catch(a){s.value=null,a instanceof Error?r.value=a:console.error(a)}finally{o.value=!1}}return(u,a)=>(e(),t(x,null,{default:m(({route:v})=>[i(B,{title:c(l)("zone-ingresses.routes.item.title",{name:v.params.zoneIngress})},null,8,["title"]),w(),i($,{breadcrumbs:[{to:{name:"zone-ingress-list-view"},text:c(l)("zone-ingresses.routes.item.breadcrumbs")}]},{default:m(()=>[z("div",A,[o.value?(e(),t(N,{key:0})):r.value!==null?(e(),t(V,{key:1,error:r.value},null,8,["error"])):s.value===null?(e(),t(E,{key:2})):(e(),b("div",C,[i(y,{"zone-ingress-overview":s.value},null,8,["zone-ingress-overview"])]))])]),_:1},8,["breadcrumbs"])]),_:1}))}});export{U as default};
