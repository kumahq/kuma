import{_ as V}from"./CodeBlock.vue_vue_type_style_index_0_lang-fa1bd6d3.js";import{d as C,r as o,o as n,i as p,w as t,j as d,a8 as v,p as m,n as c,l as r,F as _,I as g,H as N}from"./index-21079cd9.js";const S=["data-testid","innerHTML"],A=C({__name:"ConfigView",props:{data:{},notifications:{default:()=>[]}},setup(k){const l=k;function y(f){var s;const i=((s=f.zoneInsight)==null?void 0:s.subscriptions)??[];if(i.length>0){const a=i[i.length-1];if(a.config)return JSON.stringify(JSON.parse(a.config),null,2)}return null}return(f,i)=>{const s=o("RouteTitle"),a=o("KAlert"),w=o("KCard"),h=o("AppView"),b=o("RouteView");return n(),p(b,{name:"zone-cp-config-view",params:{zone:""}},{default:t(({t:u})=>[d(h,null,v({title:t(()=>[m("h2",null,[d(s,{title:u("zone-cps.routes.item.navigation.zone-cp-config-view"),render:!0},null,8,["title"])])]),default:t(()=>[c(),c(),d(w,{class:"mt-4"},{body:t(()=>[(n(!0),r(_,null,g([y(l.data)],(e,z)=>(n(),r(_,{key:z},[e!==null?(n(),p(V,{key:0,id:"code-block-zone-config",language:"json",code:e,"is-searchable":"","query-key":"zone-config"},null,8,["code"])):(n(),p(a,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:t(()=>[c(N(u("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))],64))),128))]),_:2},1024)]),_:2},[l.notifications.length>0?{name:"notifications",fn:t(()=>[m("ul",null,[(n(!0),r(_,null,g(l.notifications,e=>(n(),r("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:u(`common.warnings.${e.kind}`,e.payload)},null,8,S))),128)),c()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{A as default};
