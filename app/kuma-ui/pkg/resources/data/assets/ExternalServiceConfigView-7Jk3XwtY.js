import{_ as C}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-2-5Xpecl.js";import{E}from"./ErrorBlock-dtRm8bS3.js";import{_ as w}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-0vTtwQOF.js";import{d as R,a,o as t,b as n,w as r,e as i,m as d,f as V,p as k}from"./index-H9kuPi5I.js";import"./CodeBlock-8ft3dwz3.js";import"./CopyButton-6xMoQ2pP.js";import"./index-FZCiQto1.js";import"./toYaml-sPaYOD3i.js";import"./TextWithCopyButton-WiPvXg9-.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-_XEV9lqg.js";const Q=R({__name:"ExternalServiceConfigView",setup(y){return(S,$)=>{const _=a("RouteTitle"),m=a("DataSource"),u=a("KCard"),f=a("AppView"),g=a("RouteView");return t(),n(g,{name:"external-service-config-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:e,t:h})=>[i(f,null,{title:r(()=>[d("h2",null,[i(_,{title:h("external-services.routes.item.navigation.external-service-config-view")},null,8,["title"])])]),default:r(()=>[V(),i(m,{src:`/meshes/${e.params.mesh}/external-services/${e.params.service}`},{default:r(({data:s,error:l})=>[l?(t(),n(E,{key:0,error:l},null,8,["error"])):s===void 0?(t(),n(w,{key:1})):(t(),n(u,{key:2,"data-testid":"external-service-config"},{default:r(()=>[d("div",null,[i(C,{resource:s.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:r(({copy:o,copying:x})=>[x?(t(),n(m,{key:0,src:`/meshes/${s.mesh}/external-services/${s.name}/as/kubernetes?no-store`,onChange:c=>{o(p=>p(c))},onError:c=>{o((p,v)=>v(c))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1024))]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{Q as default};
