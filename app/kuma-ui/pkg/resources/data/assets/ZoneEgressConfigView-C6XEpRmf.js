import{d as R,e as a,o as d,k as i,w as n,a as s,b as z,j as l,ar as p,l as V}from"./index-CUmbT3FY.js";import{_ as k}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-C9nDNYfa.js";const M=R({__name:"ZoneEgressConfigView",setup(v){return(F,S)=>{const m=a("RouteTitle"),g=a("DataSource"),_=a("DataLoader"),u=a("KCard"),f=a("AppView"),C=a("RouteView");return d(),i(C,{name:"zone-egress-config-view",params:{zoneEgress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:h,uri:t})=>[s(m,{render:!1,title:h("zone-egresses.routes.item.navigation.zone-egress-config-view")},null,8,["title"]),z(),s(f,null,{default:n(()=>[s(u,null,{default:n(()=>[s(_,{src:t(l(p),"/zone-egresses/:name",{name:e.params.zoneEgress})},{default:n(({data:x})=>[s(k,{resource:x.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:n(({copy:o,copying:E})=>[E?(d(),i(g,{key:0,src:t(l(p),"/zone-egresses/:name/as/kubernetes",{name:e.params.zoneEgress},{cacheControl:"no-store"}),onChange:r=>{o(c=>c(r))},onError:r=>{o((c,w)=>w(r))}},null,8,["src","onChange","onError"])):V("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{M as default};
