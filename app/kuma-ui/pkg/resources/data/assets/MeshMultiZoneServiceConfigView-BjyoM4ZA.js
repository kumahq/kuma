import{_ as f}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-gTq31dOV.js";import{d as C,r as a,o as d,m as i,w as r,b as t,p as x}from"./index-DW4c3gZM.js";import"./CodeBlock-C9q9QC_e.js";const M=C({__name:"MeshMultiZoneServiceConfigView",props:{data:{}},setup(p){const n=p;return(w,v)=>{const m=a("DataSource"),l=a("KCard"),u=a("AppView"),_=a("RouteView");return d(),i(_,{name:"mesh-multi-zone-service-config-view",params:{mesh:"",service:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:r(({route:o})=>[t(u,null,{default:r(()=>[t(l,null,{default:r(()=>[t(f,{resource:n.data.config,"is-searchable":"",query:o.params.codeSearch,"is-filter-mode":o.params.codeFilter,"is-reg-exp-mode":o.params.codeRegExp,onQueryChange:e=>o.update({codeSearch:e}),onFilterModeChange:e=>o.update({codeFilter:e}),onRegExpModeChange:e=>o.update({codeRegExp:e})},{default:r(({copy:e,copying:h})=>[h?(d(),i(m,{key:0,src:`/meshes/${n.data.mesh}/mesh-multi-zone-service/${n.data.id}/as/kubernetes?no-store`,onChange:s=>{e(c=>c(s))},onError:s=>{e((c,g)=>g(s))}},null,8,["src","onChange","onError"])):x("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{M as default};
