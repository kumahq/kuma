import{d as f,e as n,o as c,p as m,w as r,a as p,m as C,ad as x,q as w}from"./index-CFsM3b-2.js";import{_ as E}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-04YIm1IQ.js";const F=f({__name:"DataPlaneSummaryConfigView",props:{data:{},routeName:{}},setup(d){const s=d;return(R,V)=>{const l=n("DataSource"),i=n("AppView"),u=n("RouteView");return c(),m(u,{name:s.routeName,params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:e,uri:h})=>[p(i,null,{default:r(()=>[p(E,{resource:s.data.config,language:"yaml","is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:r(({copy:a,copying:g})=>[g?(c(),m(l,{key:0,src:h(C(x),"/meshes/:mesh/dataplanes/:name/as/kubernetes",{mesh:e.params.mesh,name:e.params.dataPlane},{cacheControl:"no-store"}),onChange:o=>{a(t=>t(o))},onError:o=>{a((t,_)=>_(o))}},null,8,["src","onChange","onError"])):w("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:1},8,["name"])}}});export{F as default};
