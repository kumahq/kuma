import{d as _,L as f,o as n,g as o,w as t,h as l,p as w,A as h,m as b,C as v,i as p,l as c,a3 as z,E as $,af as k,D as q,_ as C,q as S}from"./index-18fd9432.js";import{D as V,K as x}from"./KFilterBar-597f0331.js";import"./dataplane-30467516.js";const B=_({__name:"GatewayListView",props:{page:{},size:{},search:{},query:{},mesh:{},gatewayType:{}},setup(u){const a=u,{t:g}=f();return(y,T)=>(n(),o(C,{name:"gateways-list-view"},{default:t(({route:i,can:d})=>[l(w,{src:`/meshes/${i.params.mesh}/gateways/of/${a.gatewayType}?page=${a.page}&size=${y.size}&search=${a.search}`},{default:t(({data:s,error:r})=>[l(h,null,{title:t(()=>[b("h2",null,[l(v,{title:p(g)("gateways.routes.items.title"),render:!0},null,8,["title"])])]),default:t(()=>[c(),l(p(z),null,{body:t(()=>[r!==void 0?(n(),o($,{key:0,error:r},null,8,["error"])):(n(),o(V,{key:1,"data-testid":"gateway-collection",class:"gateway-collection","page-number":a.page,"page-size":a.size,total:s==null?void 0:s.total,items:s==null?void 0:s.items,error:r,gateways:!0,onChange:({page:e,size:m})=>{i.update({page:String(e),size:String(m)})}},{toolbar:t(()=>[l(x,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:a.query,fields:{name:{description:"filter by name or parts of a name"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...d("use zones")?{zone:{description:"filter by “kuma.io/zone” value"}}:{}},onFieldsChange:e=>i.update({query:e.query,s:e.query.length>0?JSON.stringify(e.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),c(),l(p(k),{label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(e=>({...e,selected:e.value===a.gatewayType})),appearance:"select",onSelected:e=>i.update({gatewayType:String(e.value)})},{"item-template":t(({item:e})=>[c(q(e.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","onChange"]))]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:1}))}});const A=S(B,[["__scopeId","data-v-a6ddecf9"]]);export{A as default};
