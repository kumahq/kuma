import{d as u,o as _,a as f,w as t,h as s,q as d,b as o,g,G as h}from"./index-30c3bdbc.js";import{D as y,K as b}from"./KFilterBar-6713a1d8.js";import{g as z,o as q,A as v,_ as w,f as V}from"./RouteView.vue_vue_type_script_setup_true_lang-1dd6f4c1.js";import{_ as $}from"./RouteTitle.vue_vue_type_script_setup_true_lang-cbf5001a.js";import"./AppCollection-7d9e5098.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-b3f1f8ad.js";import"./dataplane-30467516.js";import"./notEmpty-7f452b20.js";const x=u({__name:"DataPlaneListView",props:{page:{},size:{},search:{},query:{},mesh:{}},setup(n){const e=n,{t:l}=z();return(p,C)=>(_(),f(w,{name:"data-planes-list-view"},{default:t(({route:r})=>[s(q,{src:`/meshes/${e.mesh}/dataplanes?page=${e.page}&size=${p.size}&search=${e.search}`},{default:t(({data:a,error:c})=>[s(v,null,{title:t(()=>[d("h2",null,[s($,{title:o(l)("data-planes.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[g(),s(o(h),null,{body:t(()=>[s(y,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":e.page,"page-size":e.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,error:c,onChange:({page:i,size:m})=>{r.update({page:String(i),size:String(m)})}},{toolbar:t(()=>[s(b,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:i=>r.update({query:i.query,s:i.query.length>0?JSON.stringify(i.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1}))}});const A=V(x,[["__scopeId","data-v-19acf073"]]);export{A as default};
