import{d as m,o as y,a as d,w as t,h as s,q as _,b as r,g as o,G as f,U as w,t as h}from"./index-7e71fe76.js";import{g as b,p as v,A as z,_ as q,f as $}from"./RouteView.vue_vue_type_script_setup_true_lang-159ad8a0.js";import{_ as S}from"./RouteTitle.vue_vue_type_script_setup_true_lang-3c1a3272.js";import{D as V,K as x}from"./KFilterBar-8c854892.js";import"./AppCollection-8d01782e.js";import"./notEmpty-7f452b20.js";const C=m({__name:"GatewayListView",props:{page:{},size:{},search:{},query:{},mesh:{},gatewayType:{}},setup(n){const a=n,{t:p}=b();return(c,T)=>(y(),d(q,{name:"gateways-list-view"},{default:t(({route:i})=>[s(v,{src:`/meshes/${i.params.mesh}/gateways/of/${a.gatewayType}?page=${a.page}&size=${c.size}&search=${a.search}`},{default:t(({data:l,error:u})=>[s(z,null,{title:t(()=>[_("h2",null,[s(S,{title:r(p)("gateways.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[o(),s(r(f),null,{body:t(()=>[s(V,{"data-testid":"gateway-collection",class:"gateway-collection","page-number":a.page,"page-size":a.size,total:l==null?void 0:l.total,items:l==null?void 0:l.items,error:u,gateways:!0,onChange:({page:e,size:g})=>{i.update({page:String(e),size:String(g)})}},{toolbar:t(()=>[s(x,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:a.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:e=>i.update({query:e.query,s:e.query.length>0?JSON.stringify(e.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),o(),s(r(w),{label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(e=>({...e,selected:e.value===a.gatewayType})),appearance:"select",onSelected:e=>i.update({gatewayType:String(e.value)})},{"item-template":t(({item:e})=>[o(h(e.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1}))}});const N=$(C,[["__scopeId","data-v-25c280de"]]);export{N as default};
