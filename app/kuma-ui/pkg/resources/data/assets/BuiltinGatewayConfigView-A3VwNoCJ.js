import{_ as y}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-CO0czkwV.js";import{d as C,i as t,o as n,a as s,w as o,j as i,g as x,k,X as R,a4 as V,e as E}from"./index-B4OAi35c.js";import"./CodeBlock-DHYtINYj.js";import"./toYaml-DB9FPXFY.js";const N=C({__name:"BuiltinGatewayConfigView",setup(v){return($,b)=>{const d=t("RouteTitle"),l=t("DataSource"),u=t("KCard"),g=t("AppView"),_=t("RouteView");return n(),s(_,{name:"builtin-gateway-config-view",params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:h})=>[i(g,null,{title:o(()=>[x("h2",null,[i(d,{title:h("builtin-gateways.routes.item.navigation.builtin-gateway-config-view")},null,8,["title"])])]),default:o(()=>[k(),i(u,null,{default:o(()=>[i(l,{src:`/meshes/${e.params.mesh}/mesh-gateways/${e.params.gateway}`},{default:o(({data:r,error:m})=>[m?(n(),s(R,{key:0,error:m},null,8,["error"])):r===void 0?(n(),s(V,{key:1})):(n(),s(y,{key:2,"data-testid":"config",resource:r.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:f})=>[f?(n(),s(l,{key:0,src:`/meshes/${r.mesh}/mesh-gateways/${r.name}/as/kubernetes?no-store`,onChange:c=>{a(p=>p(c))},onError:c=>{a((p,w)=>w(c))}},null,8,["src","onChange","onError"])):E("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{N as default};
