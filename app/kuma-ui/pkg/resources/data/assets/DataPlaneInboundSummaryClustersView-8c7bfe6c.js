import{K as C}from"./index-fce48c05.js";import{d as y,a as s,o as n,b as r,w as t,e as o,m as c,f as l,c as x,l as m,a0 as v,_ as w}from"./index-94bc0e5d.js";import{_ as R}from"./CodeBlock.vue_vue_type_style_index_0_lang-267e24b8.js";import{E as S}from"./ErrorBlock-4b0b2d3c.js";import{_ as V}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-9eb8d15b.js";import"./uniqueId-90cc9b93.js";import"./TextWithCopyButton-7a8e0cd6.js";import"./CopyButton-ff80411a.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-2fd4b808.js";const b={key:2},k={class:"toolbar"},E=y({__name:"DataPlaneInboundSummaryClustersView",setup(B){return(I,$)=>{const d=s("RouteTitle"),_=s("KButton"),u=s("DataSource"),f=s("AppView"),h=s("RouteView");return n(),r(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"data-plane-inbound-summary-clusters-view"},{default:t(({route:e})=>[o(f,null,{title:t(()=>[c("h3",null,[o(d,{title:"Clusters"})])]),default:t(()=>[l(),c("div",null,[o(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:t(({data:i,error:p,refresh:g})=>[p?(n(),r(S,{key:0,error:p},null,8,["error"])):i===void 0?(n(),r(V,{key:1})):(n(),x("div",b,[c("div",k,[o(_,{appearance:"primary",onClick:g},{default:t(()=>[o(m(v),{size:m(C)},null,8,["size"]),l(`
                  Refresh
                `)]),_:2},1032,["onClick"])]),l(),o(R,{language:"json",code:(()=>`${i.split(`
`).filter(a=>a.startsWith(`localhost:${e.params.service}::`)).map(a=>a.replace(`localhost:${e.params.service}::`,"")).join(`
`)}`)(),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});const z=w(E,[["__scopeId","data-v-6a709b8c"]]);export{z as default};
