import{d as h,r as o,o as R,m as w,w as a,b as n,e as r,p as T,U as V}from"./index-DIs9RbIP.js";const E=h({__name:"ConnectionsClustersView",props:{routeName:{}},setup(p){const c=p;return(k,t)=>{const d=o("RouteTitle"),i=o("XAction"),l=o("XCodeBlock"),m=o("DataLoader"),u=o("XCard"),_=o("AppView"),g=o("RouteView");return R(),w(g,{name:c.routeName,params:{mesh:"",proxy:"",proxyType:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:f,uri:x})=>[n(_,null,{default:a(()=>[n(d,{render:!1,title:f("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"]),t[1]||(t[1]=r()),n(u,null,{default:a(()=>[n(m,{src:x(T(V),"/connections/clusters/for/:proxyType/:name/:mesh",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*"})},{default:a(({data:C,refresh:y})=>[n(l,{language:"json",code:C,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:s=>e.update({codeSearch:s}),onFilterModeChange:s=>e.update({codeFilter:s}),onRegExpModeChange:s=>e.update({codeRegExp:s})},{"primary-actions":a(()=>[n(i,{action:"refresh",appearance:"primary",onClick:y},{default:a(()=>t[0]||(t[0]=[r(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1},8,["name"])}}});export{E as default};
