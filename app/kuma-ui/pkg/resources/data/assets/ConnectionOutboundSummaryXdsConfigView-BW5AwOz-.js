import{d as h,r as n,o as b,q as k,w as s,b as a,e as r,p as E,$ as R,t as w}from"./index-CjpuIAP7.js";const V=h({__name:"ConnectionOutboundSummaryXdsConfigView",props:{routeName:{}},setup(p){const d=p;return(S,t)=>{const i=n("RouteTitle"),l=n("XCheckbox"),m=n("XAction"),u=n("XCodeBlock"),g=n("DataLoader"),_=n("AppView"),x=n("RouteView");return b(),k(x,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,proxy:"",proxyType:"",connection:"",includeEds:Boolean},name:d.routeName},{default:s(({t:c,route:e,uri:f})=>[a(i,{render:!1,title:c("connections.routes.item.navigation.xds")},null,8,["title"]),t[1]||(t[1]=r()),a(_,null,{default:s(()=>[a(g,{src:f(E(R),"/connections/xds/for/:proxyType/:name/outbound/:outbound/endpoints/:endpoints",{name:e.params.proxy,outbound:e.params.connection,endpoints:String(e.params.includeEds),proxyType:e.params.proxyType==="ingresses"?"zone-ingress":"zone-egress"})},{default:s(({data:y,refresh:C})=>[a(u,{language:"json",code:JSON.stringify(y,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":s(()=>[a(l,{checked:e.params.includeEds,label:c("connections.include_endpoints"),onChange:o=>e.update({includeEds:o})},null,8,["checked","label","onChange"]),t[0]||(t[0]=r()),a(m,{action:"refresh",appearance:"primary",onClick:C},{default:s(()=>[r(w(c("common.refresh")),1)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{V as default};
