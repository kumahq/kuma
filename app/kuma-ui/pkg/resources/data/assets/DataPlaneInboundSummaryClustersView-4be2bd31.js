import{K as g}from"./index-fce48c05.js";import{d as x,a as t,o as n,b as r,w as o,e as s,m as p,f as i,l as m,a0 as y}from"./index-6594470f.js";import{C as R}from"./CodeBlock-13aedd48.js";import{E as k}from"./ErrorBlock-eae09ed8.js";import{_ as w}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-6c7e8a5b.js";import"./uniqueId-90cc9b93.js";import"./TextWithCopyButton-e98ec405.js";import"./CopyButton-60165c43.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-9e7cbe5c.js";const D=x({__name:"DataPlaneInboundSummaryClustersView",setup(E){return(V,v)=>{const d=t("RouteTitle"),u=t("KButton"),_=t("DataSource"),f=t("AppView"),h=t("RouteView");return n(),r(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"data-plane-inbound-summary-clusters-view"},{default:o(({route:e})=>[s(f,null,{title:o(()=>[p("h3",null,[s(d,{title:"Clusters"})])]),default:o(()=>[i(),p("div",null,[s(_,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:o(({data:c,error:l,refresh:C})=>[l?(n(),r(k,{key:0,error:l},null,8,["error"])):c===void 0?(n(),r(w,{key:1})):(n(),r(R,{key:2,language:"json",code:(()=>`${c.split(`
`).filter(a=>a.startsWith(`localhost:${e.params.service}::`)).map(a=>a.replace(`localhost:${e.params.service}::`,"")).join(`
`)}`)(),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":o(()=>[s(u,{appearance:"primary",onClick:C},{default:o(()=>[s(m(y),{size:m(g)},null,8,["size"]),i(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{D as default};
