import{m as C}from"./kong-icons.es350-D4Gbfme0.js";import{C as w}from"./CodeBlock-DCsN1aa-.js";import{d as x,i as t,o as R,a as y,w as n,j as o,g as c,k as r,A as V}from"./index-DQUwSwHF.js";const F=x({__name:"ConnectionOutboundSummaryStatsView",props:{dataplaneOverview:{}},setup(i){const p=i;return(k,v)=>{const d=t("RouteTitle"),l=t("KButton"),m=t("DataCollection"),u=t("DataLoader"),_=t("AppView"),f=t("RouteView");return R(),y(f,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-stats-view"},{default:n(({route:e})=>[o(_,null,{title:n(()=>[c("h3",null,[o(d,{title:"Stats"})])]),default:n(()=>[r(),c("div",null,[o(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${p.dataplaneOverview.dataplane.networking.inboundAddress}`},{default:n(({data:g,refresh:h})=>[o(m,{items:g.raw.split(`
`),predicate:s=>s.includes(`.${e.params.connection}.`)},{default:n(({items:s})=>[o(w,{language:"json",code:s.map(a=>a.replace(`${e.params.connection}.`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[o(l,{appearance:"primary",onClick:h},{default:n(()=>[o(V(C)),r(`

                  Refresh
                `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{F as default};
