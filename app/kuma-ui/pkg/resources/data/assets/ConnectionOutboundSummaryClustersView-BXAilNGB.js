import{d as C,e as o,o as f,m as g,w as n,a as t,b as s}from"./index-Yqc5mH7h.js";const V=C({__name:"ConnectionOutboundSummaryClustersView",setup(x){return(R,w)=>{const r=o("RouteTitle"),i=o("XAction"),p=o("XCodeBlock"),d=o("DataCollection"),l=o("DataLoader"),m=o("AppView"),_=o("RouteView");return f(),g(_,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-clusters-view"},{default:n(({route:e})=>[t(r,{render:!1,title:"Clusters"}),s(),t(m,null,{default:n(()=>[t(l,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:n(({data:u,refresh:h})=>[t(d,{items:u.split(`
`),predicate:c=>c.startsWith(`${e.params.connection}::`)},{default:n(({items:c})=>[t(p,{language:"json",code:c.map(a=>a.replace(`${e.params.connection}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[t(i,{action:"refresh",appearance:"primary",onClick:h},{default:n(()=>[s(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{V as default};
