import{_ as z}from"./CodeBlock.vue_vue_type_style_index_0_lang-2e43d83c.js";import{d as v,a as o,o as n,b as p,w as t,e as d,a5 as N,p as g,f as c,c as r,F as _,D as h,t as T}from"./index-d50afca2.js";const x=["data-testid","innerHTML"],K=v({__name:"ConfigView",props:{data:{},notifications:{default:()=>[]}},setup(y){const l=y;function k(f){var i;const a=((i=f.zoneInsight)==null?void 0:i.subscriptions)??[];if(a.length>0){const s=a[a.length-1];if(s.config)return JSON.stringify(JSON.parse(s.config),null,2)}return null}return(f,a)=>{const i=o("RouteTitle"),s=o("KAlert"),w=o("KCard"),b=o("AppView"),C=o("RouteView");return n(),p(C,{name:"zone-cp-config-view",params:{zone:"",codeSearch:""}},{default:t(({route:m,t:u})=>[d(b,null,N({title:t(()=>[g("h2",null,[d(i,{title:u("zone-cps.routes.item.navigation.zone-cp-config-view")},null,8,["title"])])]),default:t(()=>[c(),c(),d(w,null,{body:t(()=>[(n(!0),r(_,null,h([k(l.data)],(e,S)=>(n(),r(_,{key:S},[e!==null?(n(),p(z,{key:0,id:"code-block-zone-config",language:"json",code:e,"is-searchable":"",query:m.params.codeSearch,onQueryChange:V=>m.update({codeSearch:V})},null,8,["code","query","onQueryChange"])):(n(),p(s,{key:1,class:"mt-4","data-testid":"warning-no-subscriptions",appearance:"warning"},{alertMessage:t(()=>[c(T(u("zone-cps.detail.no_subscriptions")),1)]),_:2},1024))],64))),128))]),_:2},1024)]),_:2},[l.notifications.length>0?{name:"notifications",fn:t(()=>[g("ul",null,[(n(!0),r(_,null,h(l.notifications,e=>(n(),r("li",{key:e.kind,"data-testid":`warning-${e.kind}`,innerHTML:u(`common.warnings.${e.kind}`,e.payload)},null,8,x))),128)),c()])]),key:"0"}:void 0]),1024)]),_:1})}}});export{K as default};
