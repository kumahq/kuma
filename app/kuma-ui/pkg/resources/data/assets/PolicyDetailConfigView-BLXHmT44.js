import{d as C,r as o,o as t,m as p,w as n,b as c,p as y,a1 as x,q as w}from"./index-C-Llvxgw.js";import{_ as E}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-D7DuEGVs.js";const k=C({__name:"PolicyDetailConfigView",props:{data:{}},setup(i){const l=i;return(R,V)=>{const m=o("DataSource"),d=o("XCard"),h=o("AppView"),_=o("RouteView");return t(),p(_,{name:"policy-detail-config-view",params:{mesh:"",policy:"",policyPath:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,uri:u})=>[c(h,null,{default:n(()=>[c(d,null,{default:n(()=>[c(E,{resource:l.data.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:n(({copy:a,copying:f})=>[f?(t(),p(m,{key:0,src:u(y(x),"/meshes/:mesh/policy-path/:path/policy/:name/as/kubernetes",{mesh:e.params.mesh,path:e.params.policyPath,name:e.params.policy},{cacheControl:"no-cache"}),onChange:r=>{a(s=>s(r))},onError:r=>{a((s,g)=>g(r))}},null,8,["src","onChange","onError"])):w("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{k as default};
