import{d as w,e as a,o as x,m as R,w as o,a as t,b as s,l as d,a3 as k,K as V}from"./index-DpAqRYT5.js";const K=w({__name:"DataPlaneStatsView",props:{data:{}},setup(c){const r=c;return(y,B)=>{const l=a("RouteTitle"),p=a("KButton"),i=a("XCodeBlock"),m=a("DataLoader"),_=a("KCard"),u=a("AppView"),g=a("RouteView");return x(),R(g,{name:"data-plane-stats-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:f})=>[t(l,{render:!1,title:f("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"]),s(),t(u,null,{default:o(()=>[t(_,null,{default:o(()=>[t(m,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${r.data.dataplane.networking.inboundAddress}`},{default:o(({data:h,refresh:C})=>[t(i,{language:"json",code:h.raw,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":o(()=>[t(p,{appearance:"primary",onClick:C},{default:o(()=>[t(d(k),{size:d(V)},null,8,["size"]),s(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{K as default};
