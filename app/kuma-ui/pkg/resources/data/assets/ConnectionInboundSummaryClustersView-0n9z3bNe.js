import{d as g,a as n,o as t,b as c,w as o,e as s,m as i,f as p,E as x,A as y,a4 as R,q as d,K as k}from"./index-pAyRVwwQ.js";import{C as w}from"./CodeBlock-6c7dCnil.js";const $=g({__name:"ConnectionInboundSummaryClustersView",setup(E){return(V,v)=>{const m=n("RouteTitle"),u=n("KButton"),_=n("DataSource"),h=n("AppView"),f=n("RouteView");return t(),c(f,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",service:""},name:"connection-inbound-summary-clusters-view"},{default:o(({route:e})=>[s(h,null,{title:o(()=>[i("h3",null,[s(m,{title:"Clusters"})])]),default:o(()=>[p(),i("div",null,[s(_,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:o(({data:r,error:l,refresh:C})=>[l?(t(),c(x,{key:0,error:l},null,8,["error"])):r===void 0?(t(),c(y,{key:1})):(t(),c(w,{key:2,language:"json",code:`${r.split(`
`).filter(a=>a.startsWith(`localhost:${e.params.service}::`)).map(a=>a.replace(`localhost:${e.params.service}::`,"")).join(`
`)}`,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":o(()=>[s(u,{appearance:"primary",onClick:C},{default:o(()=>[s(d(R),{size:d(k)},null,8,["size"]),p(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{$ as default};
