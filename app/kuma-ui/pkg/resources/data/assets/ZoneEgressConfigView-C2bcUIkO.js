import{d as z,e as a,o as i,p,w as n,a as s,b as V,m as l,aq as m,q as v}from"./index-D8KcXOkO.js";import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-Ch_PDPX3.js";const D=z({__name:"ZoneEgressConfigView",setup(S){return(k,t)=>{const g=a("RouteTitle"),_=a("DataSource"),u=a("DataLoader"),f=a("KCard"),C=a("AppView"),h=a("RouteView");return i(),p(h,{name:"zone-egress-config-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:x,uri:c})=>[s(g,{render:!1,title:x("zone-egresses.routes.item.navigation.zone-egress-config-view")},null,8,["title"]),t[0]||(t[0]=V()),s(C,null,{default:n(()=>[s(f,null,{default:n(()=>[s(u,{src:c(l(m),"/zone-egresses/:name",{name:e.params.zoneEgress})},{default:n(({data:E})=>[s(F,{resource:E.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:n(({copy:o,copying:w})=>[w?(i(),p(_,{key:0,src:c(l(m),"/zone-egresses/:name/as/kubernetes",{name:e.params.zoneEgress},{cacheControl:"no-store"}),onChange:r=>{o(d=>d(r))},onError:r=>{o((d,R)=>R(r))}},null,8,["src","onChange","onError"])):v("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{D as default};
