import{d as h,r as n,o as k,q as w,w as a,b as t,e as p,p as R,$ as A}from"./index-BP47cGGe.js";const S=h({__name:"ConnectionOutboundSummaryStatsView",props:{networking:{},routeName:{}},setup(i){const c=i;return(T,s)=>{const d=n("RouteTitle"),m=n("XAction"),l=n("XCodeBlock"),u=n("DataCollection"),_=n("DataLoader"),g=n("AppView"),f=n("RouteView");return k(),w(f,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,proxy:"",proxyType:"",connection:""},name:c.routeName},{default:a(({route:e,uri:x})=>[t(d,{render:!1,title:"Stats"}),s[1]||(s[1]=p()),t(g,null,{default:a(()=>[t(_,{src:x(R(A),"/connections/stats/for/:proxyType/:name/:mesh/:socketAddress",{name:e.params.proxy,mesh:"*",socketAddress:c.networking.inboundAddress,proxyType:e.params.proxyType==="ingresses"?"zone-ingress":"zone-egress"})},{default:a(({data:y,refresh:C})=>[t(u,{items:y.raw.split(`
`),predicate:r=>r.includes(`.${e.params.connection}.`)},{default:a(({items:r})=>[t(l,{language:"json",code:r.map(o=>o.replace(`${e.params.connection}.`,"")).join(`
`),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":a(()=>[t(m,{action:"refresh",appearance:"primary",onClick:C},{default:a(()=>s[0]||(s[0]=[p(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{S as default};
