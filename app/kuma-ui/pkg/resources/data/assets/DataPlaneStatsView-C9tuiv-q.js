import{d as x,r as a,o as R,p as k,w as o,b as n,e as d,m as V,Z as y}from"./index-yoi81zLz.js";const F=x({__name:"DataPlaneStatsView",props:{data:{}},setup(r){const p=r;return(A,s)=>{const c=a("RouteTitle"),l=a("XAction"),i=a("XCodeBlock"),m=a("DataLoader"),_=a("KCard"),u=a("AppView"),f=a("RouteView");return R(),k(f,{name:"data-plane-stats-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:g,uri:h})=>[n(c,{render:!1,title:g("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"]),s[1]||(s[1]=d()),n(u,null,{default:o(()=>[n(_,null,{default:o(()=>[n(m,{src:h(V(y),"/meshes/:mesh/dataplanes/:name/stats/:address",{mesh:e.params.mesh,name:e.params.dataPlane,address:p.data.dataplane.networking.inboundAddress})},{default:o(({data:C,refresh:w})=>[n(i,{language:"json",code:C.raw,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":o(()=>[n(l,{action:"refresh",appearance:"primary",onClick:w},{default:o(()=>s[0]||(s[0]=[d(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{F as default};
