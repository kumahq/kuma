import{d as S,o as a,e as f,h as r,w as e,q as w,g as i,t as m,b as l,a as n,f as A,G as $,F as N,B as F,s as K,N as L}from"./index-4fccd604.js";import{m as R,g as T,D as v,S as W,R as G,A as J,o as x,p as q,E as V,_ as O,f as j}from"./RouteView.vue_vue_type_script_setup_true_lang-3d610ab4.js";import{_ as H}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-b49ea8e5.js";import{T as M}from"./TagList-a33ad563.js";import{T as k}from"./TextWithCopyButton-e620cf76.js";import{_ as Q}from"./RouteTitle.vue_vue_type_script_setup_true_lang-ac1c81c6.js";import{D as U,K as X}from"./KFilterBar-1cb850cc.js";import{T as Y}from"./TabsWidget-7dc330e7.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-66cc49af.js";import"./CopyButton-84b4f533.js";import"./AppCollection-13889b2e.js";import"./dataplane-30467516.js";import"./notEmpty-7f452b20.js";const Z={class:"stack"},D={class:"columns",style:{"--columns":"3"}},ee=S({__name:"ExternalServiceDetails",props:{serviceInsight:{},externalService:{}},setup(h){const t=h,o=R(),{t:_}=T();async function y(u){const{mesh:s,name:d}=t.externalService;return await o.getExternalService({mesh:s,name:d},u)}return(u,s)=>(a(),f("div",Z,[r(l($),null,{body:e(()=>[w("div",D,[r(v,null,{title:e(()=>[i(m(l(_)("http.api.property.name")),1)]),body:e(()=>[r(k,{text:t.serviceInsight.name},null,8,["text"])]),_:1}),i(),r(v,null,{title:e(()=>[i(m(l(_)("http.api.property.address")),1)]),body:e(()=>[i(m(t.externalService.networking.address),1)]),_:1}),i(),t.externalService.tags!==null?(a(),n(v,{key:0},{title:e(()=>[i(m(l(_)("http.api.property.tags")),1)]),body:e(()=>[r(M,{tags:t.externalService.tags},null,8,["tags"])]),_:1})):A("",!0)])]),_:1}),i(),r(H,{id:"code-block-service",resource:t.externalService,"resource-fetcher":y,"is-searchable":""},null,8,["resource"])]))}}),te={class:"stack"},se={class:"columns",style:{"--columns":"4"}},re=S({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(h){const t=h,{t:o}=T();return(_,y)=>(a(),f("div",te,[r(l($),null,{body:e(()=>{var u,s;return[w("div",se,[r(v,null,{title:e(()=>[i(m(l(o)("http.api.property.status")),1)]),body:e(()=>[r(W,{status:t.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),i(),r(v,null,{title:e(()=>[i(m(l(o)("http.api.property.name")),1)]),body:e(()=>[r(k,{text:t.serviceInsight.name},null,8,["text"])]),_:1}),i(),r(v,null,{title:e(()=>[i(m(l(o)("http.api.property.address")),1)]),body:e(()=>[t.serviceInsight.addressPort?(a(),n(k,{key:0,text:t.serviceInsight.addressPort},null,8,["text"])):(a(),f(N,{key:1},[i(m(l(o)("common.detail.none")),1)],64))]),_:1}),i(),r(G,{online:((u=t.serviceInsight.dataplanes)==null?void 0:u.online)??0,total:((s=t.serviceInsight.dataplanes)==null?void 0:s.total)??0},{title:e(()=>[i(m(l(o)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])]}),_:1})]))}}),ie=S({__name:"ServiceDetailView",props:{page:{},size:{},search:{},query:{},gatewayType:{},mesh:{},service:{}},setup(h){const t=h,{t:o}=T();function _(y){const u=[{hash:"#overview",title:o("services.routes.item.tabs.overview")}];return y.serviceType!=="external"&&u.push({hash:"#dataPlaneProxies",title:o("services.routes.item.tabs.data_plane_proxies")}),u}return(y,u)=>(a(),n(O,{name:"service-detail-view","data-testid":"service-detail-view"},{default:e(({route:s})=>[r(J,{breadcrumbs:[{to:{name:"mesh-detail-view",params:{mesh:s.params.mesh}},text:s.params.mesh},{to:{name:"services-list-view",params:{mesh:s.params.mesh}},text:l(o)("services.routes.item.breadcrumbs")}]},{title:e(()=>[w("h1",null,[r(Q,{title:l(o)("services.routes.item.title",{name:s.params.service}),render:!0},null,8,["title"])])]),default:e(()=>[i(),r(x,{src:`/meshes/${s.params.mesh}/service-insights/${s.params.service}`},{default:e(({data:d,error:I})=>[d===void 0?(a(),n(q,{key:0})):I?(a(),n(V,{key:1,error:I},null,8,["error"])):(a(),n(Y,{key:2,tabs:_(d)},F({overview:e(()=>[d.serviceType==="external"?(a(),n(x,{key:0,src:`/meshes/${s.params.mesh}/external-services/${s.params.service}`},{default:e(({data:c,error:g})=>[c===void 0?(a(),n(q,{key:0})):g?(a(),n(V,{key:1,error:g},null,8,["error"])):(a(),n(ee,{key:2,"service-insight":d,"external-service":c},null,8,["service-insight","external-service"]))]),_:2},1032,["src"])):(a(),n(re,{key:1,"service-insight":d},null,8,["service-insight"]))]),_:2},[d.serviceType!=="external"?{name:"dataPlaneProxies",fn:e(()=>[r(x,{src:`/meshes/${s.params.mesh}/dataplanes/for/${s.params.service}/of/${t.gatewayType}?page=${t.page}&size=${t.size}&search=${t.search}`},{default:e(({data:c,error:g})=>{var B,z,C,P;return[(a(!0),f(N,null,K([typeof((P=(C=(z=(B=c==null?void 0:c.items)==null?void 0:B[0])==null?void 0:z.dataplane)==null?void 0:C.networking)==null?void 0:P.gateway)>"u"],b=>(a(),n(l($),{key:b},{body:e(()=>[r(U,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":t.page,"page-size":t.size,total:c==null?void 0:c.total,items:c==null?void 0:c.items,error:g,gateways:b,onChange:({page:p,size:E})=>{s.update({page:String(p),size:String(E)})}},{toolbar:e(()=>[r(X,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:p=>s.update({query:p.query,s:p.query.length>0?JSON.stringify(p.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),i(),b?(a(),n(l(L),{key:0,label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(p=>({...p,selected:p.value===t.gatewayType})),appearance:"select",onSelected:p=>s.update({gatewayType:String(p.value)})},{"item-template":e(({item:p})=>[i(m(p.label),1)]),_:2},1032,["items","onSelected"])):A("",!0)]),_:2},1032,["page-number","page-size","total","items","error","gateways","onChange"])]),_:2},1024))),128))]}),_:2},1032,["src"])]),key:"0"}:void 0]),1032,["tabs"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});const ge=j(ie,[["__scopeId","data-v-978c7f7e"]]);export{ge as default};
