import{E as d}from"./EnvoyData-4c52d667.js";import{g as l}from"./dataplane-dcd0858b.js";import{d as g,a as e,o as _,b as f,w as o,e as t,p as h,f as w,q as C}from"./index-203d56a2.js";import"./index-9dd3e7d3.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-f7a0d7a8.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-01997eab.js";import"./ErrorBlock-085322b0.js";import"./TextWithCopyButton-45b0690a.js";import"./CopyButton-4a565fd0.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-297a38e2.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-096f0e9b.js";const b=g({__name:"XdsConfigView",props:{data:{}},setup(a){const n=a;return(V,x)=>{const r=e("RouteTitle"),i=e("KCard"),p=e("AppView"),c=e("RouteView");return _(),f(c,{name:"zone-egress-xds-config-view",params:{zoneEgress:"",codeSearch:""}},{default:o(({route:s,t:m})=>[t(p,null,{title:o(()=>[h("h2",null,[t(r,{title:m("zone-egresses.routes.item.navigation.zone-egress-xds-config-view")},null,8,["title"])])]),default:o(()=>[w(),t(i,null,{body:o(()=>[t(d,{status:C(l)(n.data.zoneEgressInsight),resource:"Zone",src:`/zone-egresses/${s.params.zoneEgress}/data-path/xds`,query:s.params.codeSearch,onQueryChange:u=>s.update({codeSearch:u})},null,8,["status","src","query","onQueryChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{b as default};
