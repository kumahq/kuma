import{d as g,a as n,o as t,b as c,w as o,e as s,m as i,f as l,E as C,A as x,a4 as y,p as R}from"./index-CP9JG8i6.js";import{C as k}from"./CodeBlock-VPwL1aP6.js";const E=g({__name:"ConnectionOutboundSummaryStatsView",setup(w){return(v,S)=>{const d=n("RouteTitle"),m=n("KButton"),u=n("DataSource"),_=n("AppView"),f=n("RouteView");return t(),c(f,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"connection-outbound-summary-stats-view"},{default:o(({route:e})=>[s(_,null,{title:o(()=>[i("h3",null,[s(d,{title:"Stats"})])]),default:o(()=>[l(),i("div",null,[s(u,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${e.params.service}`},{default:o(({data:r,error:p,refresh:h})=>[p?(t(),c(C,{key:0,error:p},null,8,["error"])):r===void 0?(t(),c(x,{key:1})):(t(),c(k,{key:2,language:"json",code:r.raw.split(`
`).filter(a=>a.includes(`.${e.params.service}.`)).map(a=>a.replace(`${e.params.service}.`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":o(()=>[s(m,{appearance:"primary",onClick:h},{default:o(()=>[s(R(y)),l(`

                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{E as default};
