import{_ as x}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-GAOsSNx9.js";import{d as k,i as n,o as t,a as s,w as o,j as r,g as w,k as R,a1 as V,a9 as E,e as y}from"./index-CyAtMQ3G.js";import"./CodeBlock-37LejceU.js";import"./toYaml-DB9FPXFY.js";const P=k({__name:"DataPlaneConfigView",setup(v){return($,F)=>{const m=n("RouteTitle"),l=n("DataSource"),_=n("KCard"),u=n("AppView"),g=n("RouteView");return t(),s(g,{name:"data-plane-config-view",params:{mesh:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:f})=>[r(u,null,{title:o(()=>[w("h2",null,[r(m,{title:f("data-planes.routes.item.navigation.data-plane-config-view")},null,8,["title"])])]),default:o(()=>[R(),r(_,null,{default:o(()=>[r(l,{src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}`},{default:o(({data:d,error:i})=>[i?(t(),s(V,{key:0,error:i},null,8,["error"])):d===void 0?(t(),s(E,{key:1})):(t(),s(x,{key:2,resource:d.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:h})=>[h?(t(),s(l,{key:0,src:`/meshes/${e.params.mesh}/dataplanes/${e.params.dataPlane}/as/kubernetes?no-store`,onChange:c=>{a(p=>p(c))},onError:c=>{a((p,C)=>C(c))}},null,8,["src","onChange","onError"])):y("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{P as default};
