import{d as x,e as a,o as y,p as w,w as n,a as t,b as d,m as R,ad as b,t as V}from"./index-CFsM3b-2.js";const v=x({__name:"ConnectionInboundSummaryXdsConfigView",props:{data:{},dataplaneOverview:{}},setup(r){const i=r;return(k,s)=>{const p=a("RouteTitle"),l=a("XAction"),m=a("XCodeBlock"),u=a("DataLoader"),_=a("AppView"),g=a("RouteView");return y(),w(g,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:""},name:"connection-inbound-summary-xds-config-view"},{default:n(({t:c,route:e,uri:f})=>[t(p,{render:!1,title:c("connections.routes.item.navigation.xds")},null,8,["title"]),s[0]||(s[0]=d()),t(_,null,{default:n(()=>[t(u,{src:f(R(b),"/meshes/:mesh/dataplanes/:dataplane/inbound/:inbound/xds",{mesh:e.params.mesh,dataplane:e.params.dataPlane,inbound:`${i.data.port}`})},{default:n(({data:h,refresh:C})=>[t(m,{language:"json",code:JSON.stringify(h,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{"primary-actions":n(()=>[t(l,{action:"refresh",appearance:"primary",onClick:C},{default:n(()=>[d(V(c("common.refresh")),1)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{v as default};
