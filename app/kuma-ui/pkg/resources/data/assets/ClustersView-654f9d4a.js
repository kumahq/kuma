import{E as m}from"./EnvoyData-c65fec9b.js";import{g as d}from"./dataplane-0a086c06.js";import{d as _,r as e,o as g,i as h,w as t,j as s,p as f,n as w,k as C}from"./index-a6f5023f.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ad731d3d.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-1974ccfb.js";const S=_({__name:"ClustersView",props:{data:{}},setup(n){const a=n;return(V,z)=>{const r=e("RouteTitle"),c=e("KCard"),i=e("AppView"),u=e("RouteView");return g(),h(u,{name:"zone-ingress-clusters-view",params:{zoneIngress:"",codeSearch:""}},{default:t(({route:o,t:p})=>[s(i,null,{title:t(()=>[f("h2",null,[s(r,{title:p("zone-ingresses.routes.item.navigation.zone-ingress-clusters-view"),render:!0},null,8,["title"])])]),default:t(()=>[w(),s(c,null,{body:t(()=>[s(m,{status:C(d)(a.data.zoneIngressInsight),resource:"Zone",src:`/zone-ingresses/${o.params.zoneIngress}/data-path/clusters`,query:o.params.codeSearch,onQueryChange:l=>o.update({codeSearch:l})},null,8,["status","src","query","onQueryChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{S as default};
