import{d as C,r as o,o as x,p as b,w as t,b as a,e as c,m as V,$ as y,t as E}from"./index-BIN9nSPF.js";const S=C({__name:"ConnectionOutboundSummaryXdsConfigView",setup(R){return(k,s)=>{const l=o("RouteTitle"),i=o("XCheckbox"),p=o("XAction"),r=o("XCodeBlock"),m=o("DataLoader"),u=o("AppView"),_=o("RouteView");return x(),b(_,{params:{codeSearch:"",codeFilter:!1,codeRegExp:!1,mesh:"",dataPlane:"",connection:"",includeEds:!1},name:"connection-outbound-summary-xds-config-view"},{default:t(({t:d,route:e,uri:g})=>[a(l,{render:!1,title:d("connections.routes.item.navigation.xds")},null,8,["title"]),s[1]||(s[1]=c()),a(u,null,{default:t(()=>[a(m,{src:g(V(y),"/meshes/:mesh/dataplanes/:dataplane/outbound/:outbound/xds/:endpoints",{mesh:e.params.mesh,dataplane:e.params.dataPlane,outbound:e.params.connection,endpoints:String(e.params.includeEds)})},{default:t(({data:f,refresh:h})=>[a(r,{language:"json",code:JSON.stringify(f,null,2),"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:n=>e.update({codeSearch:n}),onFilterModeChange:n=>e.update({codeFilter:n}),onRegExpModeChange:n=>e.update({codeRegExp:n})},{"primary-actions":t(()=>[a(i,{modelValue:e.params.includeEds,"onUpdate:modelValue":n=>e.params.includeEds=n,label:d("connections.include_endpoints")},null,8,["modelValue","onUpdate:modelValue","label"]),s[0]||(s[0]=c()),a(p,{action:"refresh",appearance:"primary",onClick:h},{default:t(()=>[c(E(d("common.refresh")),1)]),_:2},1032,["onClick"])]),_:2},1032,["code","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}});export{S as default};
