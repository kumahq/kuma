import{d as T,r as o,m as k,o as B,w as n,b as a,e as d,s as E,U as F,q as N}from"./index-C6a9CJr9.js";const v={ref:"$el"},M=T({__name:"ConnectionsClustersView",props:{routeName:{}},setup(i){const l=i;return(m,s)=>{const u=o("RouteTitle"),g=o("XAction"),_=o("XCodeBlock"),f=o("XWindow"),h=o("DataLoader"),x=o("XCard"),C=o("AppView"),y=o("RouteView");return B(),k(y,{name:l.routeName,params:{mesh:"",proxy:"",proxyType:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:R,uri:w})=>[a(C,null,{default:n(()=>[a(u,{render:!1,title:R("data-planes.routes.item.navigation.data-plane-clusters-view")},null,8,["title"]),s[1]||(s[1]=d()),a(x,null,{default:n(()=>[a(h,{src:w(E(F),"/connections/clusters/for/:proxyType/:name/:mesh",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*"})},{default:n(({data:V,refresh:X})=>[a(f,{resize:!0},{default:n(({resize:r})=>{var p,c;return[N("div",v,[a(_,{"max-height":`${(((p=r==null?void 0:r.target)==null?void 0:p.innerHeight)??0)-(((c=m.$el)==null?void 0:c.getBoundingClientRect().top)+200)}`,language:"json",code:V,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":n(()=>[a(g,{action:"refresh",appearance:"primary",onClick:X},{default:n(()=>s[0]||(s[0]=[d(`
                    Refresh
                  `)])),_:2},1032,["onClick"])]),_:2},1032,["max-height","code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])],512)]}),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1},8,["name"])}}});export{M as default};
