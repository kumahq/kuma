import{C as h}from"./CodeBlock-CzrOpKtx.js";import{d as f,r as a,o as g,m as C,w as o,b as t,e as s}from"./index-DKRUpwtt.js";const k=f({__name:"DataPlaneClustersView",setup(x){return(R,w)=>{const r=a("RouteTitle"),c=a("XAction"),d=a("DataLoader"),l=a("KCard"),p=a("AppView"),i=a("RouteView");return g(),C(i,{name:"data-plane-clusters-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:m})=>[t(p,null,{default:o(()=>[t(r,{render:!1,title:m("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"]),s(),t(l,null,{default:o(()=>[t(d,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/data-path/clusters`},{default:o(({data:_,refresh:u})=>[t(h,{language:"json",code:_,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":o(()=>[t(c,{type:"refresh",appearance:"primary",onClick:u},{default:o(()=>[s(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{k as default};
