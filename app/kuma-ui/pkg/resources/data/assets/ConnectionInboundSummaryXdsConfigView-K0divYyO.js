import{d as C,r as o,o as y,q as R,w as n,b as t,e as d,p as w,$ as V,t as b}from"./index-BP47cGGe.js";const E=C({__name:"ConnectionInboundSummaryXdsConfigView",props:{data:{},routeName:{}},setup(p){const s=p;return(k,r)=>{const i=o("RouteTitle"),l=o("XAction"),m=o("XCodeBlock"),u=o("DataLoader"),_=o("AppView"),g=o("RouteView");return y(),R(g,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",proxy:"",connection:""},name:s.routeName},{default:n(({t:c,route:e,uri:f})=>[t(i,{render:!1,title:c("connections.routes.item.navigation.xds")},null,8,["title"]),r[0]||(r[0]=d()),t(_,null,{default:n(()=>[t(u,{src:f(w(V),"/meshes/:mesh/dataplanes/:dataplane/inbound/:inbound/xds",{mesh:e.params.mesh,dataplane:e.params.proxy,inbound:`${s.data.port}`})},{default:n(({data:h,refresh:x})=>[t(m,{language:"json",code:JSON.stringify(h,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{"primary-actions":n(()=>[t(l,{action:"refresh",appearance:"primary",onClick:x},{default:n(()=>[d(b(c("common.refresh")),1)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:1},8,["name"])}}});export{E as default};
