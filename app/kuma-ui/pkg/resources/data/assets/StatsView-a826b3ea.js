import{E as l}from"./EnvoyData-c81e58ec.js";import{g as m}from"./dataplane-0a086c06.js";import{d as _,r as e,o as d,i as g,w as t,j as s,p as f,n as w,k as z}from"./index-adcc6fc8.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-4a09da2d.js";const k=_({__name:"StatsView",props:{data:{}},setup(o){const a=o;return(V,h)=>{const n=e("RouteTitle"),r=e("KCard"),i=e("AppView"),p=e("RouteView");return d(),g(p,{name:"zone-egress-stats-view",params:{zoneEgress:""}},{default:t(({route:c,t:u})=>[s(i,null,{title:t(()=>[f("h2",null,[s(n,{title:u("zone-egresses.routes.item.navigation.zone-egress-stats-view"),render:!0},null,8,["title"])])]),default:t(()=>[w(),s(r,null,{body:t(()=>[s(l,{status:z(m)(a.data.zoneEgressInsight),resource:"Zone",src:`/zone-egresses/${c.params.zoneEgress}/data-path/stats`,"query-key":"envoy-data-stats-zone-egress"},null,8,["status","src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{k as default};
