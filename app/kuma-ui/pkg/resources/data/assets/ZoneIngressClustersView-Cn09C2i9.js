import{_ as C,r as o,o as x,p as h,w as n,b as s,e as r}from"./index-BIN9nSPF.js";const R={};function w(V,t){const c=o("RouteTitle"),i=o("XAction"),d=o("XCodeBlock"),l=o("DataLoader"),p=o("XCard"),_=o("AppView"),m=o("RouteView");return x(),h(m,{name:"zone-ingress-clusters-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:g})=>[s(c,{render:!1,title:g("zone-ingresses.routes.item.navigation.zone-ingress-clusters-view")},null,8,["title"]),t[1]||(t[1]=r()),s(_,null,{default:n(()=>[s(p,null,{default:n(()=>[s(l,{src:`/zone-ingresses/${e.params.zoneIngress}/data-path/clusters`},{default:n(({data:u,refresh:f})=>[s(d,{language:"json",code:u,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[s(i,{action:"refresh",appearance:"primary",onClick:f},{default:n(()=>t[0]||(t[0]=[r(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}const y=C(R,[["render",w]]);export{y as default};
