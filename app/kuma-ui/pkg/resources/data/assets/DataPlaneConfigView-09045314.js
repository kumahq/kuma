import{d as h,R as k,r as e,o,i as n,w as a,j as s,p as w,n as V,E as C,x as g,k as v}from"./index-21079cd9.js";import{_ as x}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-343ce353.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-fa1bd6d3.js";import"./toYaml-4e00099e.js";const K=h({__name:"DataPlaneConfigView",setup(R){const l=k();return(y,A)=>{const i=e("RouteTitle"),m=e("DataSource"),p=e("KCard"),u=e("AppView"),_=e("RouteView");return o(),n(_,{name:"data-plane-config-view",params:{mesh:"",dataPlane:""}},{default:a(({route:r,t:d})=>[s(u,null,{title:a(()=>[w("h2",null,[s(i,{title:d("data-planes.routes.item.navigation.data-plane-config-view"),render:!0},null,8,["title"])])]),default:a(()=>[V(),s(p,null,{body:a(()=>[s(m,{src:`/meshes/${r.params.mesh}/dataplanes/${r.params.dataPlane}`},{default:a(({data:t,error:c})=>[c?(o(),n(C,{key:0,error:c},null,8,["error"])):t===void 0?(o(),n(g,{key:1})):(o(),n(x,{key:2,id:"code-block-data-plane",resource:t,"resource-fetcher":f=>v(l).getDataplaneFromMesh({mesh:t.mesh,name:t.name},f),"is-searchable":""},null,8,["resource","resource-fetcher"]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{K as default};
