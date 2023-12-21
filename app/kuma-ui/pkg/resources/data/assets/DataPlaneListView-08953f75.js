import{D as h,F as C}from"./FilterBar-1bee3fc7.js";import{E as V}from"./ErrorBlock-18f2aade.js";import{S as k}from"./SummaryView-76209480.js";import{d as S,a as l,o as p,b as i,w as t,e as n,p as z,f as o,t as D,D as P,s as u,_ as q}from"./index-1ce19661.js";import"./index-fce48c05.js";import"./AppCollection-ac0b1f5f.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-2d03c5f4.js";import"./StatusBadge-6636728c.js";import"./TextWithCopyButton-228c2509.js";import"./CopyButton-3ac2aa71.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-209ba38d.js";const T=S({__name:"DataPlaneListView",setup(x){return(R,B)=>{const y=l("RouteTitle"),f=l("KSelect"),g=l("KCard"),w=l("RouterView"),v=l("AppView"),m=l("DataSource"),b=l("RouteView");return p(),i(m,{src:"/me"},{default:t(({data:c})=>[c?(p(),i(b,{key:0,name:"data-plane-list-view",params:{page:1,size:c.pageSize,query:"",dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:t(({route:e,t:d})=>[n(m,{src:`/meshes/${e.params.mesh}/dataplanes/of/${e.params.dataplaneType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:t(({data:s,error:r})=>[n(v,null,{title:t(()=>[z("h2",null,[n(y,{title:d("data-planes.routes.items.title")},null,8,["title"])])]),default:t(()=>[o(),n(g,null,{default:t(()=>[r!==void 0?(p(),i(V,{key:0,error:r},null,8,["error"])):(p(),i(h,{key:1,"data-testid":"data-plane-collection","page-number":parseInt(e.params.page),"page-size":parseInt(e.params.size),total:s==null?void 0:s.total,items:s==null?void 0:s.items,error:r,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"data-plane-summary-view",onChange:e.update},{toolbar:t(()=>[n(C,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/service: backend'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),o(),n(f,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(a=>({value:a,label:d(`data-planes.type.${a}`),selected:a===e.params.dataplaneType})),appearance:"select",onSelected:a=>e.update({dataplaneType:String(a.value)})},{"item-template":t(({item:a})=>[o(D(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","is-selected-row","onChange"]))]),_:2},1024),o(),e.params.dataPlane?(p(),i(w,{key:0},{default:t(a=>[n(k,{onClose:_=>e.replace({name:"data-plane-list-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:t(()=>[(p(),i(P(a.Component),{name:e.params.dataPlane,"dataplane-overview":s==null?void 0:s.items.find(_=>_.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):u("",!0)]),_:1})}}});const G=q(T,[["__scopeId","data-v-03b837d9"]]);export{G as default};
