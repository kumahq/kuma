import{d as g,Q as h,r as e,o as r,i as n,w as o,j as a,p as k,n as w,E as y,x as z,k as C}from"./index-a6f5023f.js";import{_ as V}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-3ee102e1.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-ad731d3d.js";import"./toYaml-4e00099e.js";const b=g({__name:"ConfigView",setup(v){const u=h();return(E,x)=>{const p=e("RouteTitle"),l=e("DataSource"),_=e("KCard"),m=e("AppView"),d=e("RouteView");return r(),n(d,{name:"zone-egress-config-view",params:{zoneEgress:"",codeSearch:""}},{default:o(({route:s,t:f})=>[a(m,null,{title:o(()=>[k("h2",null,[a(p,{title:f("zone-egresses.routes.item.navigation.zone-egress-config-view"),render:!0},null,8,["title"])])]),default:o(()=>[w(),a(_,null,{body:o(()=>[a(l,{src:`/zone-egresses/${s.params.zoneEgress}`},{default:o(({data:c,error:i})=>[i!==void 0?(r(),n(y,{key:0,error:i},null,8,["error"])):c===void 0?(r(),n(z,{key:1})):(r(),n(V,{key:2,id:"code-block-zone-egress",resource:c,"resource-fetcher":t=>C(u).getZoneEgress({name:s.params.zoneEgress},t),"is-searchable":"",query:s.params.codeSearch,onQueryChange:t=>s.update({codeSearch:t})},null,8,["resource","resource-fetcher","query","onQueryChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{b as default};
