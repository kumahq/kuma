import{_ as C}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-Y2uQL_0I.js";import{d as E,a,o as n,b as s,w as r,e as c,m as d,f as w,E as R,A as V,p as k}from"./index-8scsg5Gp.js";import"./CodeBlock-O4jkqEEa.js";import"./toYaml-sPaYOD3i.js";const b=E({__name:"ExternalServiceConfigView",setup(y){return(S,$)=>{const _=a("RouteTitle"),l=a("DataSource"),u=a("KCard"),f=a("AppView"),g=a("RouteView");return n(),s(g,{name:"external-service-config-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:e,t:h})=>[c(f,null,{title:r(()=>[d("h2",null,[c(_,{title:h("external-services.routes.item.navigation.external-service-config-view")},null,8,["title"])])]),default:r(()=>[w(),c(l,{src:`/meshes/${e.params.mesh}/external-services/${e.params.service}`},{default:r(({data:t,error:m})=>[m?(n(),s(R,{key:0,error:m},null,8,["error"])):t===void 0?(n(),s(V,{key:1})):(n(),s(u,{key:2,"data-testid":"external-service-config"},{default:r(()=>[d("div",null,[c(C,{resource:t.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:r(({copy:o,copying:x})=>[x?(n(),s(l,{key:0,src:`/meshes/${t.mesh}/external-services/${t.name}/as/kubernetes?no-store`,onChange:i=>{o(p=>p(i))},onError:i=>{o((p,v)=>v(i))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1024))]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{b as default};
