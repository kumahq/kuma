import{d as y,a as t,o as s,b as l,w as a,e as o,m as x,f as p,E as R,A as E,a4 as V,q as i,K as B}from"./index-pAyRVwwQ.js";import{C as S}from"./CodeBlock-6c7dCnil.js";const N=y({__name:"DataPlaneStatsView",props:{data:{}},setup(m){const r=m;return(v,K)=>{const _=t("RouteTitle"),u=t("KButton"),g=t("DataSource"),f=t("KCard"),h=t("AppView"),C=t("RouteView");return s(),l(C,{name:"data-plane-stats-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:w})=>[o(h,null,{title:a(()=>[x("h2",null,[o(_,{title:w("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"])])]),default:a(()=>[p(),o(f,null,{default:a(()=>[o(g,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/stats/${r.data.dataplaneType==="builtin"&&r.data.dataplane.networking.gateway?r.data.dataplane.networking.gateway.tags["kuma.io/service"]:"localhost_"}`},{default:a(({data:c,error:d,refresh:k})=>[d?(s(),l(R,{key:0,error:d},null,8,["error"])):c===void 0?(s(),l(E,{key:1})):(s(),l(S,{key:2,language:"json",code:c.raw,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":a(()=>[o(u,{appearance:"primary",onClick:k},{default:a(()=>[o(i(V),{size:i(B)},null,8,["size"]),p(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{N as default};
