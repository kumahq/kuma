import{d as R,e as o,o as m,m as p,w as a,a as t,k as c,l as h,aq as f,P as x,b as n,t as _,$ as S,p as v}from"./index-CjjKwNo4.js";import{T as b}from"./TagList-pzFi9naC.js";import{_ as F}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-B26EjIdi.js";const M={class:"stack"},N={class:"columns"},A=R({__name:"ExternalServiceDetailView",setup(q){return(B,r)=>{const C=o("KCard"),y=o("DataSource"),w=o("DataLoader"),V=o("AppView"),k=o("RouteView");return m(),p(k,{name:"external-service-detail-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({route:e,t:l,uri:g})=>[t(V,null,{default:a(()=>[c("div",M,[t(w,{src:g(h(f),"/meshes/:mesh/external-services/:name",{mesh:e.params.mesh,name:e.params.service})},{default:a(({data:i})=>[t(C,{"data-testid":"external-service-details"},{default:a(()=>[c("div",N,[t(x,null,{title:a(()=>[n(_(l("http.api.property.address")),1)]),body:a(()=>[t(S,{text:i.networking.address},null,8,["text"])]),_:2},1024),r[2]||(r[2]=n()),i.tags?(m(),p(x,{key:0},{title:a(()=>[n(_(l("http.api.property.tags")),1)]),body:a(()=>[t(b,{tags:i.tags},null,8,["tags"])]),_:2},1024)):v("",!0)])]),_:2},1024),r[4]||(r[4]=n()),c("div",null,[c("h3",null,_(l("external-services.detail.config")),1),r[3]||(r[3]=n()),t(F,{class:"mt-4","data-testid":"external-service-config",resource:i.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:s=>e.update({codeSearch:s}),onFilterModeChange:s=>e.update({codeFilter:s}),onRegExpModeChange:s=>e.update({codeRegExp:s})},{default:a(({copy:s,copying:E})=>[E?(m(),p(y,{key:0,src:g(h(f),"/meshes/:mesh/external-services/:name/as/kubernetes",{mesh:e.params.mesh,name:e.params.service},{cacheControl:"no-store"}),onChange:d=>{s(u=>u(d))},onError:d=>{s((u,D)=>D(d))}},null,8,["src","onChange","onError"])):v("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])])]),_:2},1032,["src"])])]),_:2},1024)]),_:1})}}});export{A as default};
