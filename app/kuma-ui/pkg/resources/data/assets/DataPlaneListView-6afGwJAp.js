import{D as C,F as V}from"./FilterBar-7MFWgv28.js";import{d as k,a as l,o as i,b as o,w as s,e as n,m as z,f as p,E as S,t as D,D as q,p as y,_ as P}from"./index-eDf0gHDD.js";import{S as T}from"./SummaryView-FjMJCvcI.js";import"./AppCollection-ReyhGgL1.js";import"./StatusBadge-HXpDGqsA.js";const x=k({__name:"DataPlaneListView",setup(R){return(B,N)=>{const g=l("RouteTitle"),f=l("KSelect"),w=l("KCard"),b=l("RouterView"),h=l("AppView"),m=l("DataSource"),v=l("RouteView");return i(),o(m,{src:"/me"},{default:s(({data:c})=>[c?(i(),o(v,{key:0,name:"data-plane-list-view",params:{page:1,size:c.pageSize,query:"",dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:s(({can:d,route:e,t:u})=>[n(m,{src:`/meshes/${e.params.mesh}/dataplanes/of/${e.params.dataplaneType}?page=${e.params.page}&size=${e.params.size}&search=${e.params.s}`},{default:s(({data:t,error:r})=>[n(h,null,{title:s(()=>[z("h2",null,[n(g,{title:u("data-planes.routes.items.title")},null,8,["title"])])]),default:s(()=>[p(),n(w,null,{default:s(()=>[r!==void 0?(i(),o(S,{key:0,error:r},null,8,["error"])):(i(),o(C,{key:1,"data-testid":"data-plane-collection","page-number":e.params.page,"page-size":e.params.size,total:t==null?void 0:t.total,items:t==null?void 0:t.items,error:r,"is-selected-row":a=>a.name===e.params.dataPlane,"summary-route-name":"data-plane-summary-view","is-global-mode":d("use zones"),"can-use-gateways-ui":d("use gateways ui"),onChange:e.update},{toolbar:s(()=>[n(V,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/service: backend'",query:e.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:a=>e.update({query:a.query,s:a.query.length>0?JSON.stringify(a.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),p(),n(f,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(a=>({value:a,label:u(`data-planes.type.${a}`),selected:a===e.params.dataplaneType})),onSelected:a=>e.update({dataplaneType:String(a.value)})},{"item-template":s(({item:a})=>[p(D(a.label),1)]),_:2},1032,["items","onSelected"])]),_:2},1032,["page-number","page-size","total","items","error","is-selected-row","is-global-mode","can-use-gateways-ui","onChange"]))]),_:2},1024),p(),e.params.dataPlane?(i(),o(b,{key:0},{default:s(a=>[n(T,{onClose:_=>e.replace({name:"data-plane-list-view",params:{mesh:e.params.mesh},query:{page:e.params.page,size:e.params.size}})},{default:s(()=>[(i(),o(q(a.Component),{name:e.params.dataPlane,"dataplane-overview":t==null?void 0:t.items.find(_=>_.name===e.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):y("",!0)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):y("",!0)]),_:1})}}}),E=P(x,[["__scopeId","data-v-1059cddc"]]);export{E as default};
