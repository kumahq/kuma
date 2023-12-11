import{E as z}from"./ErrorBlock-a3710a04.js";import{D as C,F as V}from"./FilterBar-2aa65b49.js";import{S as x}from"./SummaryView-39651fdf.js";import{d as b,a as r,o as i,b as n,w as s,e as o,p as P,f as p,t as k,D as q,s as u,_ as T}from"./index-e9fbefd3.js";import"./index-fce48c05.js";import"./TextWithCopyButton-0bfc7306.js";import"./CopyButton-6f1494f2.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-060e8475.js";import"./index-3d038f44.js";import"./AppCollection-98909f49.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-15e0e5b5.js";import"./StatusBadge-494c559b.js";const R=b({__name:"ServiceDataPlaneProxiesView",setup($){return(B,N)=>{const f=r("RouteTitle"),y=r("KSelect"),g=r("KCard"),v=r("RouterView"),m=r("DataSource"),w=r("AppView"),h=r("RouteView");return i(),n(m,{src:"/me"},{default:s(({data:c})=>[c?(i(),n(h,{key:0,name:"service-data-plane-proxies-view",params:{page:1,size:c.pageSize,query:"",dataplaneType:"all",s:"",mesh:"",service:"",dataPlane:""}},{default:s(({can:S,route:e,t:d})=>[o(w,null,{title:s(()=>[P("h2",null,[o(f,{title:d("services.routes.item.navigation.service-data-plane-proxies-view")},null,8,["title"])])]),default:s(()=>[p(),o(m,{src:`/meshes/${e.params.mesh}/dataplanes/for/${e.params.service}/of/${e.params.dataplaneType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:s(({data:t,error:l})=>[o(g,null,{default:s(()=>[l!==void 0?(i(),n(z,{key:0,error:l},null,8,["error"])):(i(),n(C,{key:1,"data-testid":"data-plane-collection","page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:l,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","can-use-zones":S("use zones"),onChange:e.update},{toolbar:s(()=>[o(V,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),p(),o(y,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(a=>({value:a,label:d(`data-planes.type.${a}`),selected:a===e.params.dataplaneType})),appearance:"select",onSelected:a=>e.update({dataplaneType:String(a.value)})},{"item-template":s(({item:a})=>[p(k(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","is-selected-row","can-use-zones","onChange"]))]),_:2},1024),p(),e.params.dataPlane?(i(),n(v,{key:0},{default:s(a=>[o(x,{onClose:_=>e.replace({name:"service-data-plane-proxies-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),n(q(a.Component),{name:e.params.dataPlane,"dataplane-overview":t==null?void 0:t.items.find(_=>_.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):u("",!0)]),_:1})}}});const Q=T(R,[["__scopeId","data-v-03f166d0"]]);export{Q as default};
