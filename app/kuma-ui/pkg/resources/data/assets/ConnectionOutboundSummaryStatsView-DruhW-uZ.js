import{d as w,e as n,o as x,p as R,w as o,a as t,b as p}from"./index-CFsM3b-2.js";const V=w({__name:"ConnectionOutboundSummaryStatsView",props:{dataplaneOverview:{}},setup(r){const i=r;return(y,s)=>{const d=n("RouteTitle"),l=n("XAction"),m=n("XCodeBlock"),_=n("DataCollection"),u=n("DataLoader"),f=n("AppView"),g=n("RouteView");return x(),R(g,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-stats-view"},{default:o(({route:e})=>[t(d,{render:!1,title:"Stats"}),s[1]||(s[1]=p()),t(f,null,{default:o(()=>[t(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${i.dataplaneOverview.dataplane.networking.inboundAddress}`},{default:o(({data:C,refresh:h})=>[t(_,{items:C.raw.split(`
`),predicate:c=>c.includes(`.${e.params.connection}.`)},{default:o(({items:c})=>[t(m,{language:"json",code:c.map(a=>a.replace(`${e.params.connection}.`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":o(()=>[t(l,{action:"refresh",appearance:"primary",onClick:h},{default:o(()=>s[0]||(s[0]=[p(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{V as default};
