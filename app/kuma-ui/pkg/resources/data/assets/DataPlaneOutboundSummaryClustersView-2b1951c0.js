import{K as C}from"./index-fce48c05.js";import{d as y,a as t,o as n,b as r,w as s,e as o,m as c,f as l,c as x,l as m,a0 as v,_ as w}from"./index-94bc0e5d.js";import{_ as R}from"./CodeBlock.vue_vue_type_style_index_0_lang-267e24b8.js";import{E as S}from"./ErrorBlock-4b0b2d3c.js";import{_ as V}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-9eb8d15b.js";import"./uniqueId-90cc9b93.js";import"./TextWithCopyButton-7a8e0cd6.js";import"./CopyButton-ff80411a.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-2fd4b808.js";const b={key:2},k={class:"toolbar"},E=y({__name:"DataPlaneOutboundSummaryClustersView",setup(B){return($,D)=>{const d=t("RouteTitle"),_=t("KButton"),u=t("DataSource"),f=t("AppView"),h=t("RouteView");return n(),r(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"data-plane-outbound-summary-clusters-view"},{default:s(({route:e})=>[o(f,null,{title:s(()=>[c("h3",null,[o(d,{title:"Clusters"})])]),default:s(()=>[l(),c("div",null,[o(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:s(({data:i,error:p,refresh:g})=>[p?(n(),r(S,{key:0,error:p},null,8,["error"])):i===void 0?(n(),r(V,{key:1})):(n(),x("div",b,[c("div",k,[o(_,{appearance:"primary",onClick:g},{default:s(()=>[o(m(v),{size:m(C)},null,8,["size"]),l(`
                  Refresh
                `)]),_:2},1032,["onClick"])]),l(),o(R,{language:"json",code:(()=>`${i.split(`
`).filter(a=>a.startsWith(`${e.params.service}::`)).map(a=>a.replace(`${e.params.service}::`,"")).join(`
`)}`)(),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});const q=w(E,[["__scopeId","data-v-1eb15a8f"]]);export{q as default};
