import{d as z,e as n,o as i,m as d,w as a,a as s,b as V,l,au as p,p as E}from"./index-C4IVBmnO.js";import{_ as v}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-DxUeq031.js";const I=z({__name:"ZoneIngressConfigView",setup(F){return(S,k)=>{const m=n("RouteTitle"),g=n("DataSource"),_=n("DataLoader"),u=n("KCard"),f=n("AppView"),C=n("RouteView");return i(),d(C,{name:"zone-ingress-config-view",params:{zoneIngress:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:h,uri:t})=>[s(m,{render:!1,title:h("zone-ingresses.routes.item.navigation.zone-ingress-config-view")},null,8,["title"]),V(),s(f,null,{default:a(()=>[s(u,null,{default:a(()=>[s(_,{src:t(l(p),"/zone-ingresses/:name",{name:e.params.zoneIngress})},{default:a(({data:x})=>[s(v,{resource:x.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:o=>e.update({codeSearch:o}),onFilterModeChange:o=>e.update({codeFilter:o}),onRegExpModeChange:o=>e.update({codeRegExp:o})},{default:a(({copy:o,copying:w})=>[w?(i(),d(g,{key:0,src:t(l(p),"/zone-ingresses/:name/as/kubernetes",{name:e.params.zoneIngress},{cacheControl:"no-store"}),onChange:r=>{o(c=>c(r))},onError:r=>{o((c,R)=>R(r))}},null,8,["src","onChange","onError"])):E("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{I as default};
