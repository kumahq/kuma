import{d as C,r as t,o as g,m as x,w as n,b as o,e as r,l as R,au as w}from"./index-DKRUpwtt.js";import{C as y}from"./CodeBlock-CzrOpKtx.js";const F=C({__name:"ConnectionInboundSummaryClustersView",props:{data:{}},setup(d){const c=d;return(V,k)=>{const p=t("RouteTitle"),i=t("KButton"),l=t("DataCollection"),m=t("DataLoader"),u=t("AppView"),_=t("RouteView");return g(),x(_,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-inbound-summary-clusters-view"},{default:n(({route:e})=>[o(p,{render:!1,title:"Clusters"}),r(),o(u,null,{default:n(()=>[o(m,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:n(({data:f,refresh:h})=>[o(l,{items:f.split(`
`),predicate:s=>s.startsWith(`${c.data.service}::`)},{default:n(({items:s})=>[o(y,{language:"json",code:s.map(a=>a.replace(`${c.data.service}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[o(i,{appearance:"primary",onClick:h},{default:n(()=>[o(R(w)),r(`

                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{F as default};
