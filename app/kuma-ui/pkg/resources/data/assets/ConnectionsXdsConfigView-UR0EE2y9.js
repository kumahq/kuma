import{d as k,r as n,o as E,q as R,w as t,b as a,e as p,p as X,$ as w}from"./index-CjpuIAP7.js";const b=k({__name:"ConnectionsXdsConfigView",props:{routeName:{}},setup(r){const d=r;return(T,s)=>{const c=n("RouteTitle"),i=n("XCheckbox"),l=n("XAction"),m=n("XCodeBlock"),g=n("DataLoader"),u=n("XCard"),_=n("AppView"),f=n("RouteView");return E(),R(f,{name:d.routeName,params:{mesh:"",proxy:"",proxyType:"",codeSearch:"",codeFilter:!1,codeRegExp:!1,includeEds:Boolean}},{default:t(({route:e,t:x,uri:C})=>[a(c,{render:!1,title:x("data-planes.routes.item.navigation.data-plane-xds-config-view")},null,8,["title"]),s[2]||(s[2]=p()),a(_,null,{default:t(()=>[a(u,null,{default:t(()=>[a(g,{src:C(X(w),"/connections/xds/for/:proxyType/:name/:mesh/:endpoints",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*",endpoints:String(e.params.includeEds)})},{default:t(({data:h,refresh:y})=>[a(m,{language:"json",code:JSON.stringify(h,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":t(()=>[a(i,{checked:e.params.includeEds,label:"Include Endpoints",onChange:o=>e.update({includeEds:o})},null,8,["checked","onChange"]),s[1]||(s[1]=p()),a(l,{action:"refresh",appearance:"primary",onClick:y},{default:t(()=>s[0]||(s[0]=[p(`
                Refresh
              `)])),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1},8,["name"])}}});export{b as default};
