import{d as g,e as o,o as x,m as R,w as t,a as n,b as r}from"./index-Yqc5mH7h.js";const k=g({__name:"ConnectionInboundSummaryClustersView",props:{data:{}},setup(d){const s=d;return(w,y)=>{const i=o("RouteTitle"),p=o("XAction"),l=o("XCodeBlock"),m=o("DataCollection"),_=o("DataLoader"),u=o("AppView"),h=o("RouteView");return x(),R(h,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-inbound-summary-clusters-view"},{default:t(({route:e})=>[n(i,{render:!1,title:"Clusters"}),r(),n(u,null,{default:t(()=>[n(_,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:t(({data:C,refresh:f})=>[n(m,{items:C.split(`
`),predicate:c=>c.startsWith(`${s.data.service}::`)},{default:t(({items:c})=>[n(l,{language:"json",code:c.map(a=>a.replace(`${s.data.service}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":t(()=>[n(p,{action:"refresh",appearance:"primary",onClick:f},{default:t(()=>[r(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{k as default};
