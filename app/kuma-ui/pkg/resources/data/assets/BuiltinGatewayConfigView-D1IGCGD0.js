import{d as R,r as o,o as m,m as i,w as t,b as n,e as E,l,aE as p,p as V}from"./index-BEf7lh5F.js";import{_ as b}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-BHcCkEzW.js";import"./CodeBlock-4KznwjAD.js";import"./toYaml-DB9FPXFY.js";const N=R({__name:"BuiltinGatewayConfigView",setup(v){return(F,S)=>{const d=o("RouteTitle"),g=o("DataSource"),_=o("DataLoader"),u=o("KCard"),h=o("AppView"),f=o("RouteView");return m(),i(f,{name:"builtin-gateway-config-view",params:{mesh:"",gateway:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:t(({route:e,t:w,uri:r})=>[n(d,{render:!1,title:w("builtin-gateways.routes.item.navigation.builtin-gateway-config-view")},null,8,["title"]),E(),n(h,null,{default:t(()=>[n(u,null,{default:t(()=>[n(_,{src:r(l(p),"/meshes/:mesh/mesh-gateways/:name",{mesh:e.params.mesh,name:e.params.gateway})},{default:t(({data:C})=>[n(b,{"data-testid":"config",resource:C.config,"is-searchable":"",query:e.params.codeSearch,"is-filter-mode":e.params.codeFilter,"is-reg-exp-mode":e.params.codeRegExp,onQueryChange:a=>e.update({codeSearch:a}),onFilterModeChange:a=>e.update({codeFilter:a}),onRegExpModeChange:a=>e.update({codeRegExp:a})},{default:t(({copy:a,copying:y})=>[y?(m(),i(g,{key:0,src:r(l(p),"/meshes/:mesh/mesh-gateways/:name/as/kubernetes",{mesh:e.params.mesh,name:e.params.gateway},{cacheControl:"no-store"}),onChange:s=>{a(c=>c(s))},onError:s=>{a((c,x)=>x(s))}},null,8,["src","onChange","onError"])):V("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{N as default};
