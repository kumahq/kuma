import{d as X,r as o,m as A,o as T,w as n,b as a,e as i,s as B,U as E,q as F}from"./index-De9sDLbp.js";const N={ref:"$el"},M=X({__name:"ConnectionsStatsView",props:{networking:{},routeName:{}},setup(l){const d=l;return(m,s)=>{const g=o("RouteTitle"),_=o("XAction"),u=o("XCodeBlock"),f=o("XWindow"),h=o("DataLoader"),x=o("XCard"),C=o("AppView"),y=o("RouteView");return T(),A(y,{name:d.routeName,params:{mesh:"",proxy:"",proxyType:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:w,uri:R})=>[a(g,{render:!1,title:w("data-planes.routes.item.navigation.data-plane-stats-view")},null,8,["title"]),s[1]||(s[1]=i()),a(C,null,{default:n(()=>[a(x,null,{default:n(()=>[a(h,{src:R(B(E),"/connections/stats/for/:proxyType/:name/:mesh/:socketAddress",{proxyType:{ingresses:"zone-ingress",egresses:"zone-egress"}[e.params.proxyType]??"dataplane",name:e.params.proxy,mesh:e.params.mesh||"*",socketAddress:d.networking.inboundAddress})},{default:n(({data:k,refresh:V})=>[a(f,{resize:!0},{default:n(({resize:r})=>{var p,c;return[F("div",N,[a(u,{"max-height":`${(((p=r==null?void 0:r.target)==null?void 0:p.innerHeight)??0)-(((c=m.$el)==null?void 0:c.getBoundingClientRect().top)+200)}`,language:"json",code:k.raw,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:t=>e.update({codeSearch:t}),onFilterModeChange:t=>e.update({codeFilter:t}),onRegExpModeChange:t=>e.update({codeRegExp:t})},{"primary-actions":n(()=>[a(_,{action:"refresh",appearance:"primary",onClick:V},{default:n(()=>s[0]||(s[0]=[i(`
                    Refresh
                  `)])),_:2},1032,["onClick"])]),_:2},1032,["max-height","code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])],512)]}),_:2},1024)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1},8,["name"])}}});export{M as default};
