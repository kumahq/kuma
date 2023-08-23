import{d as y,o as _,a as d,w as t,h as s,q as f,b as r,g as o,G as w,N as b,t as h}from"./index-0105b00b.js";import{g as v,o as z,A as q,_ as $,f as S}from"./RouteView.vue_vue_type_script_setup_true_lang-fff04e01.js";import{_ as V}from"./RouteTitle.vue_vue_type_script_setup_true_lang-d728f45f.js";import{D as x,K as C}from"./KFilterBar-f24e451f.js";import"./AppCollection-930d9b16.js";import"./dataplane-30467516.js";import"./notEmpty-7f452b20.js";const T=y({__name:"GatewayListView",props:{page:{},size:{},search:{},query:{},mesh:{},gatewayType:{}},setup(n){const a=n,{t:p}=v();return(c,k)=>(_(),d($,{name:"gateways-list-view"},{default:t(({route:i,can:u})=>[s(z,{src:`/meshes/${i.params.mesh}/gateways/of/${a.gatewayType}?page=${a.page}&size=${c.size}&search=${a.search}`},{default:t(({data:l,error:g})=>[s(q,null,{title:t(()=>[f("h2",null,[s(V,{title:r(p)("gateways.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[o(),s(r(w),null,{body:t(()=>[s(x,{"data-testid":"gateway-collection",class:"gateway-collection","page-number":a.page,"page-size":a.size,total:l==null?void 0:l.total,items:l==null?void 0:l.items,error:g,gateways:!0,onChange:({page:e,size:m})=>{i.update({page:String(e),size:String(m)})}},{toolbar:t(()=>[s(C,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:a.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...u("use zones")?{zone:{description:"filter by “kuma.io/zone” value"}}:{}},onFieldsChange:e=>i.update({query:e.query,s:e.query.length>0?JSON.stringify(e.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),o(),s(r(b),{label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(e=>({...e,selected:e.value===a.gatewayType})),appearance:"select",onSelected:e=>i.update({gatewayType:String(e.value)})},{"item-template":t(({item:e})=>[o(h(e.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1}))}});const I=S(T,[["__scopeId","data-v-aacbc830"]]);export{I as default};
