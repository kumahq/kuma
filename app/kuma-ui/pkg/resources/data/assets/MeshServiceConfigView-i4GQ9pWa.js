import{_ as f}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BDFzF6sj.js";import{d as C,r as a,o as d,q as i,w as r,b as t,s as x}from"./index-CKQWVGYP.js";const V=C({__name:"MeshServiceConfigView",props:{data:{}},setup(p){const n=p;return(w,v)=>{const m=a("DataSource"),l=a("XCard"),_=a("AppView"),u=a("RouteView");return d(),i(u,{name:"mesh-service-config-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:o})=>[t(_,null,{default:r(()=>[t(l,null,{default:r(()=>[t(f,{resource:n.data.config,"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:e=>o.update({codeSearch:e}),onFilterModeChange:e=>o.update({codeFilter:e}),onRegExpModeChange:e=>o.update({codeRegExp:e})},{default:r(({copy:e,copying:h})=>[h?(d(),i(m,{key:0,src:`/meshes/${n.data.mesh}/mesh-service/${n.data.id}/as/kubernetes?no-store`,onChange:s=>{e(c=>c(s))},onError:s=>{e((c,g)=>g(s))}},null,8,["src","onChange","onError"])):x("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{V as default};
