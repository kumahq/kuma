import{_ as y}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-owWKimIX.js";import{d as C,a as t,o as n,b as s,w as o,e as i,m as x,f as E,E as R,A as V,p as k}from"./index-1j9z4Egf.js";import"./CodeBlock-ejECcgv-.js";import"./toYaml-sPaYOD3i.js";const N=C({__name:"BuiltinGatewayConfigView",setup(b){return(v,$)=>{const d=t("RouteTitle"),l=t("DataSource"),u=t("KCard"),_=t("AppView"),g=t("RouteView");return n(),s(g,{name:"builtin-gateway-config-view",params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:o(({route:e,t:f})=>[i(_,null,{title:o(()=>[x("h2",null,[i(d,{title:f("builtin-gateways.routes.item.navigation.builtin-gateway-config-view")},null,8,["title"])])]),default:o(()=>[E(),i(u,null,{default:o(()=>[i(l,{src:`/meshes/${e.params.mesh}/mesh-gateways/${e.params.gateway}`},{default:o(({data:r,error:m})=>[m?(n(),s(R,{key:0,error:m},null,8,["error"])):r===void 0?(n(),s(V,{key:1})):(n(),s(y,{key:2,"data-testid":"config",resource:r.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:o(({copy:a,copying:h})=>[h?(n(),s(l,{key:0,src:`/meshes/${r.mesh}/mesh-gateways/${r.name}/as/kubernetes?no-store`,onChange:c=>{a(p=>p(c))},onError:c=>{a((p,w)=>w(c))}},null,8,["src","onChange","onError"])):k("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{N as default};
