import{d as g,e as o,o as h,m as x,w as n,a as t,b as r}from"./index-DpJ_igul.js";const V=g({__name:"ConnectionOutboundSummaryClustersView",setup(R){return(w,s)=>{const i=o("RouteTitle"),p=o("XAction"),d=o("XCodeBlock"),l=o("DataCollection"),m=o("DataLoader"),_=o("AppView"),u=o("RouteView");return h(),x(u,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-outbound-summary-clusters-view"},{default:n(({route:e})=>[t(i,{render:!1,title:"Clusters"}),s[1]||(s[1]=r()),t(_,null,{default:n(()=>[t(m,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:n(({data:C,refresh:f})=>[t(l,{items:C.split(`
`),predicate:c=>c.startsWith(`${e.params.connection}::`)},{default:n(({items:c})=>[t(d,{language:"json",code:c.map(a=>a.replace(`${e.params.connection}::`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[t(p,{action:"refresh",appearance:"primary",onClick:f},{default:n(()=>s[0]||(s[0]=[r(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{V as default};
