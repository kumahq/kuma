import{d as C,r as o,o as A,q as k,w as t,b as s,e as c,p as w,$ as R}from"./index-DRTqvjTb.js";const V=C({__name:"ConnectionInboundSummaryStatsView",props:{data:{},networking:{},routeName:{}},setup(p){const e=p;return($,d)=>{const i=o("RouteTitle"),l=o("XAction"),m=o("XCodeBlock"),g=o("DataCollection"),_=o("DataLoader"),u=o("AppView"),f=o("RouteView");return A(),k(f,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",proxy:"",proxyType:"",connection:""},name:e.routeName},{default:t(({route:a,uri:h})=>[s(i,{render:!1,title:"Stats"}),d[1]||(d[1]=c()),s(u,null,{default:t(()=>[s(_,{src:h(w(R),"/connections/stats/for/:proxyType/:name/:mesh/:socketAddress",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[a.params.proxyType]??"dataplane",name:a.params.proxy,mesh:a.params.mesh||"*",socketAddress:e.networking.inboundAddress})},{default:t(({data:x,refresh:y})=>[s(g,{items:x.raw.split(`
`),predicate:r=>[`listener.${e.data.listenerAddress.length>0?e.data.listenerAddress:a.params.connection}`,`cluster.${e.data.name}.`,`http.${e.data.name}.`,`tcp.${e.data.name}.`].some(n=>r.startsWith(n))&&(!r.includes(".rds.")||r.includes(`_${e.data.port}`))},{default:t(({items:r})=>[s(m,{language:"json",code:r.map(n=>n.replace(`${e.data.listenerAddress.length>0?e.data.listenerAddress:a.params.connection}.`,"").replace(`${e.data.name}.`,"")).join(`
`),"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:n=>a.update({codeSearch:n}),onFilterModeChange:n=>a.update({codeFilter:n}),onRegExpModeChange:n=>a.update({codeRegExp:n})},{"primary-actions":t(()=>[s(l,{action:"refresh",appearance:"primary",onClick:y},{default:t(()=>d[0]||(d[0]=[c(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["items","predicate"])]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{V as default};
