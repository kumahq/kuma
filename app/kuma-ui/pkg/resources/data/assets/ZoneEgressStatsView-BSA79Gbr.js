import{q as C,e as o,o as x,m as h,w as a,a as n,b as r}from"./index-CKcsX_-l.js";const R={};function w(E,s){const c=o("RouteTitle"),d=o("XAction"),i=o("XCodeBlock"),p=o("DataLoader"),l=o("KCard"),m=o("AppView"),_=o("RouteView");return x(),h(_,{name:"zone-egress-stats-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:g})=>[n(c,{render:!1,title:g("zone-egresses.routes.item.navigation.zone-egress-stats-view")},null,8,["title"]),s[1]||(s[1]=r()),n(m,null,{default:a(()=>[n(l,null,{default:a(()=>[n(p,{src:`/zone-egresses/${e.params.zoneEgress}/data-path/stats`},{default:a(({data:f,refresh:u})=>[n(i,{language:"json",code:f,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":a(()=>[n(d,{action:"refresh",appearance:"primary",onClick:u},{default:a(()=>s[0]||(s[0]=[r(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}const k=C(R,[["render",w]]);export{k as default};
