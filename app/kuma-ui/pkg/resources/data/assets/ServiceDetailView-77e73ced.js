import{d as x,r as V,o as a,e as k,h as s,w as e,q as S,g as r,t as u,b as n,a as m,f as L,G as w,B as R,F as E,s as F,N as K}from"./index-bb115be3.js";import{m as W,g as $,D as _,S as G,R as J,A as O,o as b,p as q,E as P,_ as j,f as H}from"./RouteView.vue_vue_type_script_setup_true_lang-e50bfc04.js";import{_ as M}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-73008343.js";import{T as Q}from"./TagList-37ab1d4c.js";import{T as A}from"./TextWithCopyButton-8bd76c4b.js";import{_ as U}from"./RouteTitle.vue_vue_type_script_setup_true_lang-c0548d11.js";import{D as X,K as Y}from"./KFilterBar-b7635a49.js";import{T as Z}from"./TabsWidget-4ef098a4.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-bc7de660.js";import"./CopyButton-9e08aa37.js";import"./AppCollection-57392b4b.js";import"./dataplane-30467516.js";const D={class:"stack"},ee={class:"columns",style:{"--columns":"3"}},te=x({__name:"ExternalServiceDetails",props:{serviceInsight:{},externalService:{}},setup(h){const t=h,l=W(),{t:v}=$();async function g(d){const{mesh:i,name:o}=t.externalService;return await l.getExternalService({mesh:i,name:o},d)}return(d,i)=>{const o=V("RouterLink");return a(),k("div",D,[s(n(w),null,{body:e(()=>[S("div",ee,[s(_,null,{title:e(()=>[r(u(n(v)("http.api.property.name")),1)]),body:e(()=>[s(A,{text:t.serviceInsight.name},{default:e(()=>[s(o,{to:{name:"service-detail-view",params:{service:t.serviceInsight.name,mesh:t.serviceInsight.mesh}}},{default:e(()=>[r(u(t.serviceInsight.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),_:1}),r(),s(_,null,{title:e(()=>[r(u(n(v)("http.api.property.address")),1)]),body:e(()=>[r(u(t.externalService.networking.address),1)]),_:1}),r(),t.externalService.tags!==null?(a(),m(_,{key:0},{title:e(()=>[r(u(n(v)("http.api.property.tags")),1)]),body:e(()=>[s(Q,{tags:t.externalService.tags},null,8,["tags"])]),_:1})):L("",!0)])]),_:1}),r(),s(M,{id:"code-block-service",resource:t.externalService,"resource-fetcher":g,"is-searchable":""},null,8,["resource"])])}}}),se={class:"stack"},re={class:"columns",style:{"--columns":"4"}},ie=x({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(h){const t=h,{t:l}=$();return(v,g)=>{const d=V("RouterLink");return a(),k("div",se,[s(n(w),null,{body:e(()=>{var i,o;return[S("div",re,[s(_,null,{title:e(()=>[r(u(n(l)("http.api.property.status")),1)]),body:e(()=>[s(G,{status:t.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),r(),s(_,null,{title:e(()=>[r(u(n(l)("http.api.property.name")),1)]),body:e(()=>[s(A,{text:t.serviceInsight.name},{default:e(()=>[s(d,{to:{name:"service-detail-view",params:{service:t.serviceInsight.name,mesh:t.serviceInsight.mesh}}},{default:e(()=>[r(u(t.serviceInsight.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),_:1}),r(),s(_,null,{title:e(()=>[r(u(n(l)("http.api.property.address")),1)]),body:e(()=>[r(u(t.serviceInsight.addressPort??n(l)("common.detail.none")),1)]),_:1}),r(),s(J,{online:((i=t.serviceInsight.dataplanes)==null?void 0:i.online)??0,total:((o=t.serviceInsight.dataplanes)==null?void 0:o.total)??0},{title:e(()=>[r(u(n(l)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])]}),_:1})])}}}),ae=x({__name:"ServiceDetailView",props:{page:{},size:{},search:{},query:{},gatewayType:{},mesh:{},service:{}},setup(h){const t=h,{t:l}=$();function v(g){const d=[{hash:"#overview",title:l("services.routes.item.tabs.overview")}];return g.serviceType==="internal"&&d.push({hash:"#dataPlaneProxies",title:l("services.routes.item.tabs.data_plane_proxies")}),d}return(g,d)=>(a(),m(j,{name:"service-detail-view","data-testid":"service-detail-view"},{default:e(({route:i})=>[s(O,{breadcrumbs:[{to:{name:"services-list-view",params:{mesh:i.params.mesh}},text:n(l)("services.routes.item.breadcrumbs")}]},{title:e(()=>[S("h2",null,[s(U,{title:n(l)("services.routes.item.title",{name:i.params.service}),render:!0},null,8,["title"])])]),default:e(()=>[r(),s(b,{src:`/meshes/${i.params.mesh}/service-insights/${i.params.service}`},{default:e(({data:o,error:I})=>[o===void 0?(a(),m(q,{key:0})):I?(a(),m(P,{key:1,error:I},null,8,["error"])):(a(),m(Z,{key:2,tabs:v(o)},R({overview:e(()=>[o.serviceType==="external"?(a(),m(b,{key:0,src:`/meshes/${i.params.mesh}/external-services/${i.params.service}`},{default:e(({data:c,error:y})=>[c===void 0?(a(),m(q,{key:0})):y?(a(),m(P,{key:1,error:y},null,8,["error"])):(a(),m(te,{key:2,"service-insight":o,"external-service":c},null,8,["service-insight","external-service"]))]),_:2},1032,["src"])):(a(),m(ie,{key:1,"service-insight":o},null,8,["service-insight"]))]),_:2},[o.serviceType!=="external"?{name:"dataPlaneProxies",fn:e(()=>[s(b,{src:`/meshes/${i.params.mesh}/dataplanes/for/${i.params.service}/of/${t.gatewayType}?page=${t.page}&size=${t.size}&search=${t.search}`},{default:e(({data:c,error:y})=>{var T,B,C,z;return[(a(!0),k(E,null,F([typeof((z=(C=(B=(T=c==null?void 0:c.items)==null?void 0:T[0])==null?void 0:B.dataplane)==null?void 0:C.networking)==null?void 0:z.gateway)>"u"],f=>(a(),m(n(w),{key:f},{body:e(()=>[s(X,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":t.page,"page-size":t.size,total:c==null?void 0:c.total,items:c==null?void 0:c.items,error:y,gateways:f,onChange:({page:p,size:N})=>{i.update({page:String(p),size:String(N)})}},{toolbar:e(()=>[s(Y,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:p=>i.update({query:p.query,s:p.query.length>0?JSON.stringify(p.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),r(),f?(a(),m(n(K),{key:0,label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(p=>({...p,selected:p.value===t.gatewayType})),appearance:"select",onSelected:p=>i.update({gatewayType:String(p.value)})},{"item-template":e(({item:p})=>[r(u(p.label),1)]),_:2},1032,["items","onSelected"])):L("",!0)]),_:2},1032,["page-number","page-size","total","items","error","gateways","onChange"])]),_:2},1024))),128))]}),_:2},1032,["src"])]),key:"0"}:void 0]),1032,["tabs"]))]),_:2},1032,["src"])]),_:2},1032,["breadcrumbs"])]),_:1}))}});const ye=H(ae,[["__scopeId","data-v-27632fde"]]);export{ye as default};
