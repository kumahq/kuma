import{d as C,a as t,o as f,b as g,w as o,e as n,m as c,f as r,p as x,ae as R}from"./index-BRR4OZXP.js";import{C as w}from"./CodeBlock-hG5z8uUD.js";const F=C({__name:"ConnectionOutboundSummaryClustersView",setup(y){return(V,B)=>{const p=t("RouteTitle"),i=t("KButton"),d=t("DataCollection"),l=t("DataLoader"),m=t("AppView"),u=t("RouteView");return f(),g(u,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-clusters-view"},{default:o(({route:e})=>[n(m,null,{title:o(()=>[c("h3",null,[n(p,{title:"Clusters"})])]),default:o(()=>[r(),c("div",null,[n(l,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:o(({data:_,refresh:h})=>[n(d,{items:_.split(`
`),predicate:s=>s.startsWith(`${e.params.connection}::`)},{default:o(({items:s})=>[n(w,{language:"json",code:s.map(a=>a.replace(`${e.params.connection}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":o(()=>[n(i,{appearance:"primary",onClick:h},{default:o(()=>[n(x(R)),r(`

                  Refresh
                `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{F as default};
