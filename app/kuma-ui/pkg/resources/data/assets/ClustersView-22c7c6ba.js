import{E as p}from"./EnvoyData-500188cc.js";import{g as m}from"./dataplane-0a086c06.js";import{d as _,r as e,o as d,g,w as t,h as s,m as f,l as w,i as z}from"./index-213666ad.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-9634fc68.js";const x=_({__name:"ClustersView",props:{data:{}},setup(o){const r=o;return(V,h)=>{const n=e("RouteTitle"),a=e("KCard"),u=e("AppView"),c=e("RouteView");return d(),g(c,{name:"zone-egress-clusters-view",params:{zoneEgress:""}},{default:t(({route:l,t:i})=>[s(u,null,{title:t(()=>[f("h2",null,[s(n,{title:i("zone-egresses.routes.item.navigation.zone-egress-clusters-view"),render:!0},null,8,["title"])])]),default:t(()=>[w(),s(a,null,{body:t(()=>[s(p,{status:z(m)(r.data.zoneEgressInsight),resource:"Zone",src:`/zone-egresses/${l.params.zoneEgress}/data-path/clusters`,"query-key":"envoy-data-clusters-zone-egress"},null,8,["status","src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{x as default};
