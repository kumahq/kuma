import{d as V,e as o,o as i,m as l,w as n,a as t,b as E,l as d,an as p,p as b}from"./index-CjjKwNo4.js";import{_ as v}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-B26EjIdi.js";const M=V({__name:"BuiltinGatewayConfigView",setup(F){return(S,r)=>{const g=o("RouteTitle"),u=o("DataSource"),_=o("DataLoader"),h=o("KCard"),f=o("AppView"),w=o("RouteView");return i(),l(w,{name:"builtin-gateway-config-view",params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:n(({route:e,t:C,uri:c})=>[t(g,{render:!1,title:C("builtin-gateways.routes.item.navigation.builtin-gateway-config-view")},null,8,["title"]),r[0]||(r[0]=E()),t(f,null,{default:n(()=>[t(h,null,{default:n(()=>[t(_,{src:c(d(p),"/meshes/:mesh/mesh-gateways/:name",{mesh:e.params.mesh,name:e.params.gateway})},{default:n(({data:y})=>[t(v,{"data-testid":"config",resource:y.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:n(({copy:a,copying:x})=>[x?(i(),l(u,{key:0,src:c(d(p),"/meshes/:mesh/mesh-gateways/:name/as/kubernetes",{mesh:e.params.mesh,name:e.params.gateway},{cacheControl:"no-store"}),onChange:s=>{a(m=>m(s))},onError:s=>{a((m,R)=>R(s))}},null,8,["src","onChange","onError"])):b("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{M as default};
