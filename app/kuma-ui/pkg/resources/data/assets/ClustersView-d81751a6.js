import{E as p}from"./EnvoyData-3228efeb.js";import{g as m}from"./dataplane-0a086c06.js";import{d as _,r as e,o as d,i as g,w as t,j as s,p as f,n as w,k as z}from"./index-78eccadf.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-13487ae9.js";const k=_({__name:"ClustersView",props:{data:{}},setup(n){const o=n;return(V,h)=>{const r=e("RouteTitle"),a=e("KCard"),i=e("AppView"),u=e("RouteView");return d(),g(u,{name:"zone-ingress-clusters-view",params:{zoneIngress:""}},{default:t(({route:c,t:l})=>[s(i,null,{title:t(()=>[f("h2",null,[s(r,{title:l("zone-ingresses.routes.item.navigation.zone-ingress-clusters-view"),render:!0},null,8,["title"])])]),default:t(()=>[w(),s(a,null,{body:t(()=>[s(p,{status:z(m)(o.data.zoneIngressInsight),resource:"Zone",src:`/zone-ingresses/${c.params.zoneIngress}/data-path/clusters`,"query-key":"envoy-data-clusters-zone-ingress"},null,8,["status","src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{k as default};
