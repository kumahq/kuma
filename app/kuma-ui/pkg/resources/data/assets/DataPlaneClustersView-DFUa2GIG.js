import{d as h,r as a,o as x,p as R,w as o,b as n,e as r,m as w,a0 as V}from"./index-Ds8TyCyC.js";const E=h({__name:"DataPlaneClustersView",setup(k){return(y,s)=>{const c=a("RouteTitle"),l=a("XAction"),d=a("XCodeBlock"),p=a("DataLoader"),i=a("XCard"),m=a("AppView"),u=a("RouteView");return x(),R(u,{name:"data-plane-clusters-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:_,uri:f})=>[n(m,null,{default:o(()=>[n(c,{render:!1,title:_("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"]),s[1]||(s[1]=r()),n(i,null,{default:o(()=>[n(p,{src:f(w(V),"/meshes/:mesh/dataplanes/:name/clusters",{mesh:e.params.mesh,name:e.params.dataPlane})},{default:o(({data:g,refresh:C})=>[n(d,{language:"json",code:g,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":o(()=>[n(l,{action:"refresh",appearance:"primary",onClick:C},{default:o(()=>s[0]||(s[0]=[r(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{E as default};
