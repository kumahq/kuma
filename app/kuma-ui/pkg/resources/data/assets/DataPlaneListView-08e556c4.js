import{d as u,o as m,a as d,w as t,h as s,s as f,b as n,g,j as y}from"./index-2a9ba339.js";import{D as h,K as b}from"./KFilterBar-5b962dd6.js";import{g as v,A as z,_ as w,f as V}from"./RouteView.vue_vue_type_script_setup_true_lang-5d6806ed.js";import{_ as q}from"./DataSource.vue_vue_type_script_setup_true_lang-c18fdd22.js";import{_ as x}from"./RouteTitle.vue_vue_type_script_setup_true_lang-4859e7c4.js";import"./StatusBadge-4403951c.js";const $=u({__name:"DataPlaneListView",props:{page:{},size:{},search:{},query:{},mesh:{}},setup(o){const e=o,{t:l}=v();return(p,C)=>(m(),d(w,{name:"data-planes-list-view"},{default:t(({route:r})=>[s(q,{src:`/${e.mesh}/dataplanes?page=${e.page}&size=${p.size}&search=${e.search}`},{default:t(({data:a,error:c})=>[s(z,null,{title:t(()=>[f("h2",null,[s(x,{title:n(l)("data-planes.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[g(),s(n(y),null,{body:t(()=>[s(h,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":e.page,"page-size":e.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,error:c,onChange:({page:i,size:_})=>{r.update({page:String(i),size:String(_)})}},{toolbar:t(()=>[s(b,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:i=>r.update({query:i.query,s:i.query.length>0?JSON.stringify(i.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1}))}});const F=V($,[["__scopeId","data-v-48755d39"]]);export{F as default};
