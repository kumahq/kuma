import{E as z}from"./ErrorBlock-60126425.js";import{D as C,F as V}from"./FilterBar-23a8a285.js";import{S as b}from"./SummaryView-9a3817bd.js";import{d as x,a as t,o as i,b as n,w as s,e as o,p as P,f as p,t as k,G as q,s as u,_ as T}from"./index-b149f888.js";import"./index-52545d1d.js";import"./TextWithCopyButton-e76d5e92.js";import"./CopyButton-2c5a3630.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-0cd9b9fd.js";import"./index-f3a1108a.js";import"./AppCollection-7d262a83.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-52d4c52a.js";import"./StatusBadge-cda559a5.js";import"./TagList-8c58a129.js";import"./dataplane-c0570285.js";const R=x({__name:"ServiceDataPlaneProxiesView",setup($){return(B,N)=>{const y=t("RouteTitle"),f=t("KSelect"),g=t("KCard"),v=t("RouterView"),m=t("DataSource"),w=t("AppView"),h=t("RouteView");return i(),n(m,{src:"/me"},{default:s(({data:c})=>[c?(i(),n(h,{key:0,name:"service-data-plane-proxies-view",params:{page:1,size:c.pageSize,query:"",dataplaneType:"all",s:"",mesh:"",service:"",dataPlane:""}},{default:s(({can:S,route:e,t:d})=>[o(w,null,{title:s(()=>[P("h2",null,[o(y,{title:d("services.routes.item.navigation.service-data-plane-proxies-view")},null,8,["title"])])]),default:s(()=>[p(),o(m,{src:`/meshes/${e.params.mesh}/dataplanes/for/${e.params.service}/of/${e.params.dataplaneType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:s(({data:r,error:l})=>[o(g,null,{body:s(()=>[l!==void 0?(i(),n(z,{key:0,error:l},null,8,["error"])):(i(),n(C,{key:1,"data-testid":"data-plane-collection","page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:r==null?void 0:r.total,items:r==null?void 0:r.items,error:l,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","can-use-zones":S("use zones"),onChange:e.update},{toolbar:s(()=>[o(V,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),p(),o(f,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(a=>({value:a,label:d(`data-planes.type.${a}`),selected:a===e.params.dataplaneType})),appearance:"select",onSelected:a=>e.update({dataplaneType:String(a.value)})},{"item-template":s(({item:a})=>[p(k(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","is-selected-row","can-use-zones","onChange"]))]),_:2},1024),p(),e.params.dataPlane?(i(),n(v,{key:0},{default:s(a=>[o(b,{onClose:_=>e.replace({name:"service-data-plane-proxies-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),n(q(a.Component),{name:e.params.dataPlane,"dataplane-overview":r==null?void 0:r.items.find(_=>_.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):u("",!0)]),_:1})}}});const W=T(R,[["__scopeId","data-v-c3018f71"]]);export{W as default};
