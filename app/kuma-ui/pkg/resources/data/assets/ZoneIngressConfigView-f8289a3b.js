import{d as g,N as w,r as e,o as s,g as n,w as o,h as r,m as k,l as z,E as V,s as h,i as v}from"./index-65a641bf.js";import{_ as C}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-49e3a69b.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-54c405e7.js";import"./toYaml-4e00099e.js";const b=g({__name:"ZoneIngressConfigView",setup(I){const i=w();return(x,y)=>{const l=e("RouteTitle"),m=e("DataSource"),u=e("KCard"),_=e("AppView"),p=e("RouteView");return s(),n(p,{name:"zone-ingress-config-view",params:{zoneIngress:""}},{default:o(({route:t,t:d})=>[r(_,null,{title:o(()=>[k("h2",null,[r(l,{title:d("zone-ingresses.routes.item.navigation.zone-ingress-config-view"),render:!0},null,8,["title"])])]),default:o(()=>[z(),r(u,{class:"mt-4"},{body:o(()=>[r(m,{src:`/zone-ingresses/${t.params.zoneIngress}`},{default:o(({data:a,error:c})=>[c!==void 0?(s(),n(V,{key:0,error:c},null,8,["error"])):a===void 0?(s(),n(h,{key:1})):(s(),n(C,{key:2,id:"code-block-zone-ingress",resource:a,"resource-fetcher":f=>v(i).getZoneIngress({name:t.params.zoneIngress},f),"is-searchable":""},null,8,["resource","resource-fetcher"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{b as default};
