import{d as w,e as a,o as x,m as R,w as o,a as t,b as d}from"./index-B_icS-nL.js";const y=w({__name:"DataPlaneStatsView",props:{data:{}},setup(r){const p=r;return(k,s)=>{const c=a("RouteTitle"),l=a("XAction"),i=a("XCodeBlock"),m=a("DataLoader"),_=a("KCard"),u=a("AppView"),g=a("RouteView");return x(),R(g,{name:"data-plane-stats-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:f})=>[t(c,{render:!1,title:f("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"]),s[1]||(s[1]=d()),t(u,null,{default:o(()=>[t(_,null,{default:o(()=>[t(m,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${p.data.dataplane.networking.inboundAddress}`},{default:o(({data:C,refresh:h})=>[t(i,{language:"json",code:C.raw,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":o(()=>[t(l,{action:"refresh",appearance:"primary",onClick:h},{default:o(()=>s[0]||(s[0]=[d(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{y as default};
