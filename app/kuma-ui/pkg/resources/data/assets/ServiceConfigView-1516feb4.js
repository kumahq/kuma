import{d as v,R as k,r as a,o,i as t,w as e,j as r,p as m,n as w,E as g,x as V,af as y,H as C,k as R}from"./index-78eccadf.js";import{_ as $}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-74215da3.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-13487ae9.js";import"./toYaml-4e00099e.js";const T=v({__name:"ServiceConfigView",setup(x){const l=k();return(A,B)=>{const _=a("RouteTitle"),p=a("DataSource"),u=a("KCard"),d=a("AppView"),f=a("RouteView");return o(),t(f,{name:"service-config-view",params:{mesh:"",service:""}},{default:e(({route:n,t:c})=>[r(d,null,{title:e(()=>[m("h2",null,[r(_,{title:c("services.routes.item.navigation.service-config-view"),render:!0},null,8,["title"])])]),default:e(()=>[w(),r(u,null,{body:e(()=>[r(p,{src:`/meshes/${n.params.mesh}/external-services/for/${n.params.service}`},{default:e(({data:s,error:i})=>[i?(o(),t(g,{key:0,error:i},null,8,["error"])):s===void 0?(o(),t(V,{key:1})):s===null?(o(),t(y,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[m("p",null,C(c("services.detail.no_matching_external_service",{name:n.params.service})),1)]),_:2},1024)):(o(),t($,{key:3,id:"code-block-service",resource:s,"resource-fetcher":h=>R(l).getExternalService({mesh:s.mesh,name:s.name},h),"is-searchable":""},null,8,["resource","resource-fetcher"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{T as default};
