import{q as u,e as o,o as h,m as C,w as a,a as n,b as s}from"./index-C4IVBmnO.js";const x={};function R(w,E){const c=o("RouteTitle"),r=o("XAction"),d=o("XCodeBlock"),i=o("DataLoader"),p=o("KCard"),l=o("AppView"),_=o("RouteView");return h(),C(_,{name:"zone-egress-stats-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:m})=>[n(c,{render:!1,title:m("zone-egresses.routes.item.navigation.zone-egress-stats-view")},null,8,["title"]),s(),n(l,null,{default:a(()=>[n(p,null,{default:a(()=>[n(i,{src:`/zone-egresses/${e.params.zoneEgress}/data-path/stats`},{default:a(({data:g,refresh:f})=>[n(d,{language:"json",code:g,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":a(()=>[n(r,{action:"refresh",appearance:"primary",onClick:f},{default:a(()=>[s(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}const k=u(x,[["render",R]]);export{k as default};
