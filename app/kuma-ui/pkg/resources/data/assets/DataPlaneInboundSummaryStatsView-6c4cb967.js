import{K as y}from"./index-fce48c05.js";import{d as S,a as t,o as s,b as r,w as n,e as o,m as c,f as i,c as x,l as d,a0 as C,_ as v}from"./index-12461953.js";import{_ as w}from"./CodeBlock.vue_vue_type_style_index_0_lang-cb530502.js";import{E as R}from"./ErrorBlock-b60858eb.js";import{_ as V}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-6454f6ef.js";import"./uniqueId-90cc9b93.js";import"./TextWithCopyButton-9a13e075.js";import"./CopyButton-7dcbb455.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-9289ba78.js";const b={key:2},k={class:"toolbar"},E=S({__name:"DataPlaneInboundSummaryStatsView",setup(B){return(I,$)=>{const m=t("RouteTitle"),_=t("KButton"),u=t("DataSource"),f=t("AppView"),h=t("RouteView");return s(),r(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"data-plane-inbound-summary-stats-view"},{default:n(({route:e})=>[o(f,null,{title:n(()=>[c("h3",null,[o(m,{title:"Stats"})])]),default:n(()=>[i(),c("div",null,[o(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/stats`},{default:n(({data:l,error:p,refresh:g})=>[p?(s(),r(R,{key:0,error:p},null,8,["error"])):l===void 0?(s(),r(V,{key:1})):(s(),x("div",b,[c("div",k,[o(_,{appearance:"primary",onClick:g},{default:n(()=>[o(d(C),{size:d(y)},null,8,["size"]),i(`
                  Refresh
                `)]),_:2},1032,["onClick"])]),i(),o(w,{language:"json",code:`${l.split(`
`).filter(a=>a.includes(`.localhost_${e.params.service}.`)).join(`
`)}`,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},null,8,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});const z=v(E,[["__scopeId","data-v-35a43b4a"]]);export{z as default};
