import{_ as v}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-448d7af6.js";import{E as g}from"./ErrorBlock-1fa583ae.js";import{_ as y}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-fa4c5616.js";import{_ as k}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-1528cfbc.js";import{d as w,u as C,a,o as s,b as t,w as e,e as c,p as i,f as V,t as $,q as b}from"./index-079a3a85.js";import"./index-52545d1d.js";import"./TextWithCopyButton-f3080f30.js";import"./CopyButton-86a7f09c.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b68734c9.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-e7d3cf8e.js";import"./toYaml-4e00099e.js";const F=w({__name:"ServiceConfigView",setup(A){const l=C();return(B,R)=>{const _=a("RouteTitle"),u=a("DataSource"),d=a("KCard"),f=a("AppView"),h=a("RouteView");return s(),t(h,{name:"service-config-view",params:{mesh:"",service:"",codeSearch:""}},{default:e(({route:o,t:m})=>[c(f,null,{title:e(()=>[i("h2",null,[c(_,{title:m("services.routes.item.navigation.service-config-view")},null,8,["title"])])]),default:e(()=>[V(),c(d,null,{body:e(()=>[i("div",null,[c(u,{src:`/meshes/${o.params.mesh}/external-services/for/${o.params.service}`},{default:e(({data:r,error:p})=>[p?(s(),t(g,{key:0,error:p},null,8,["error"])):r===void 0?(s(),t(y,{key:1})):r===null?(s(),t(v,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[i("p",null,$(m("services.detail.no_matching_external_service",{name:o.params.service})),1)]),_:2},1024)):(s(),t(k,{key:3,id:"code-block-service",resource:r,"resource-fetcher":n=>b(l).getExternalService({mesh:r.mesh,name:r.name},n),"is-searchable":"",query:o.params.codeSearch,onQueryChange:n=>o.update({codeSearch:n})},null,8,["resource","resource-fetcher","query","onQueryChange"]))]),_:2},1032,["src"])])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{F as default};
