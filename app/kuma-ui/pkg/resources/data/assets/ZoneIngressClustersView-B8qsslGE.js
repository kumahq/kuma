import{C as u}from"./CodeBlock-Cjpgpd0X.js";import{d as f,r as o,o as C,m as h,w as n,b as a,e as t}from"./index-JFoySG5Y.js";const z=f({__name:"ZoneIngressClustersView",setup(x){return(R,w)=>{const r=o("RouteTitle"),c=o("XAction"),i=o("DataLoader"),d=o("KCard"),p=o("AppView"),l=o("RouteView");return C(),h(l,{name:"zone-ingress-clusters-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:m})=>[a(r,{render:!1,title:m("zone-ingresses.routes.item.navigation.zone-ingress-clusters-view")},null,8,["title"]),t(),a(p,null,{default:n(()=>[a(d,null,{default:n(()=>[a(i,{src:`/zone-ingresses/${e.params.zoneIngress}/data-path/clusters`},{default:n(({data:_,refresh:g})=>[a(u,{language:"json",code:_,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:s=>e.update({codeSearch:s}),onFilterModeChange:s=>e.update({codeFilter:s}),onRegExpModeChange:s=>e.update({codeRegExp:s})},{"primary-actions":n(()=>[a(c,{action:"refresh",appearance:"primary",onClick:g},{default:n(()=>[t(`
                Refresh
              `)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{z as default};
