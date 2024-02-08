import{E as C}from"./ErrorBlock-cX4HzFOd.js";import{D as S,F as V}from"./FilterBar-mqFaQQPJ.js";import{S as z}from"./SummaryView-6_tz1MrP.js";import{d as x,a as r,o as i,b as l,w as s,e as t,m as k,f as n,t as P,C as q,p as u,_ as T}from"./index-Mzgj6Y4G.js";import"./index-FZCiQto1.js";import"./TextWithCopyButton-1wEvk7ih.js";import"./CopyButton-TPQGf4OQ.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-uxLaaMfP.js";import"./AppCollection-yVQ84eJP.js";import"./StatusBadge-JZQHHWZm.js";const R=x({__name:"ServiceDataPlaneProxiesView",setup($){return(B,N)=>{const f=r("RouteTitle"),y=r("KSelect"),g=r("KCard"),v=r("RouterView"),p=r("DataSource"),w=r("AppView"),h=r("RouteView");return i(),l(p,{src:"/me"},{default:s(({data:c})=>[c?(i(),l(h,{key:0,name:"service-data-plane-proxies-view",params:{page:1,size:c.pageSize,query:"",dataplaneType:"all",s:"",mesh:"",service:"",dataPlane:""}},{default:s(({can:b,route:e,t:d})=>[t(w,null,{title:s(()=>[k("h2",null,[t(f,{title:d("services.routes.item.navigation.service-data-plane-proxies-view")},null,8,["title"])])]),default:s(()=>[n(),t(p,{src:`/meshes/${e.params.mesh}/dataplanes/for/${e.params.service}/of/${e.params.dataplaneType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:s(({data:o,error:m})=>[t(g,null,{default:s(()=>[m!==void 0?(i(),l(C,{key:0,error:m},null,8,["error"])):(i(),l(S,{key:1,"data-testid":"data-plane-collection","page-number":e.params.page,"page-size":e.params.size,total:o==null?void 0:o.total,items:o==null?void 0:o.items,error:m,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","is-global-mode":b("use zones"),onChange:e.update},{toolbar:s(()=>[t(V,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),n(),t(y,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(a=>({value:a,label:d(`data-planes.type.${a}`),selected:a===e.params.dataplaneType})),onSelected:a=>e.update({dataplaneType:String(a.value)})},{"item-template":s(({item:a})=>[n(P(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","is-selected-row","is-global-mode","onChange"]))]),_:2},1024),n(),e.params.dataPlane?(i(),l(v,{key:0},{default:s(a=>[t(z,{onClose:_=>e.replace({name:"service-data-plane-proxies-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),l(q(a.Component),{name:e.params.dataPlane,"dataplane-overview":o==null?void 0:o.items.find(_=>_.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):u("",!0)]),_:1})}}}),H=T(R,[["__scopeId","data-v-daea0c6b"]]);export{H as default};
