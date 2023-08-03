import{d as L,c as f,r as A,o as r,a as i,w as a,q as h,g as t,h as o,t as x,f as y,e as $,F as q,b as T,G as F,j as k,s as K,L as O}from"./index-a928d02c.js";import{D as j,a as w}from"./DefinitionListItem-e5ff9064.js";import{_ as G}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-b05d43b6.js";import{S as J}from"./StatusBadge-de159c5b.js";import{T as R}from"./TagList-a2d26691.js";import{T as W}from"./TextWithCopyButton-406f01b5.js";import{m as E,f as I,g as H,A as M,q as Q,_ as U}from"./RouteView.vue_vue_type_script_setup_true_lang-f622f9ae.js";import{_ as X}from"./DataSource.vue_vue_type_script_setup_true_lang-8eeb0b78.js";import{_ as Y}from"./RouteTitle.vue_vue_type_script_setup_true_lang-a99b1649.js";import{_ as Z}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-be8f92ee.js";import{E as ee}from"./ErrorBlock-57542f11.js";import{D as te,K as re}from"./KFilterBar-eae51d87.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-8fce21bd.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-67f31cc3.js";import"./AppCollection-1deef537.js";import"./notEmpty-7f452b20.js";const ae={class:"entity-section-list"},se={class:"entity-title"},le=L({__name:"ServiceSummary",props:{service:{},externalService:{default:null}},setup(D){const e=D,g=E(),S=f(()=>({name:"service-detail-view",params:{service:e.service.name,mesh:e.service.mesh}})),u=f(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.networking.address:e.service.addressPort??null),d=f(()=>{var s;return e.service.serviceType==="external"&&e.externalService!==null?(s=e.externalService.networking.tls)!=null&&s.enabled?"Enabled":"Disabled":null}),_=f(()=>{var s,c;if(e.service.serviceType==="external")return null;{const l=((s=e.service.dataplanes)==null?void 0:s.online)??0,m=((c=e.service.dataplanes)==null?void 0:c.total)??0;return`${l} online / ${m} total`}}),p=f(()=>e.service.serviceType==="external"?null:e.service.status??null),b=f(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.tags:null);async function C(s){if(e.service.serviceType==="external"&&e.externalService!==null){const{mesh:c,name:l}=e.externalService;return await g.getExternalService({mesh:c,name:l},s)}else{const{mesh:c,name:l}=e.service;return await g.getServiceInsight({mesh:c,name:l},s)}}return(s,c)=>{const l=A("router-link");return r(),i(T(F),null,{body:a(()=>[h("div",ae,[h("section",null,[h("h1",se,[h("span",null,[t(`
              Service:

              `),o(W,{text:e.service.name},{default:a(()=>[o(l,{to:S.value},{default:a(()=>[t(x(e.service.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),t(),p.value?(r(),i(J,{key:0,status:p.value},null,8,["status"])):y("",!0)]),t(),o(j,{class:"mt-4"},{default:a(()=>[o(w,{term:"Address"},{default:a(()=>[u.value!==null?(r(),$(q,{key:0},[t(x(u.value),1)],64)):(r(),$(q,{key:1},[t(`
                —
              `)],64))]),_:1}),t(),d.value!==null?(r(),i(w,{key:0,term:"TLS"},{default:a(()=>[t(x(d.value),1)]),_:1})):y("",!0),t(),_.value!==null?(r(),i(w,{key:1,term:"Data Plane Proxies"},{default:a(()=>[t(x(_.value),1)]),_:1})):y("",!0),t(),b.value!==null?(r(),i(w,{key:2,term:"Tags"},{default:a(()=>[o(R,{tags:b.value},null,8,["tags"])]),_:1})):y("",!0)]),_:1})]),t(),o(G,{id:"code-block-service","resource-fetcher":C,"resource-fetcher-watch-key":e.service.name,"is-searchable":"","show-copy-as-kubernetes-button":e.service.serviceType==="external"&&e.externalService!==null,"code-max-height":"250px"},null,8,["resource-fetcher-watch-key","show-copy-as-kubernetes-button"])])]),_:1})}}});const ne=I(le,[["__scopeId","data-v-31d05cbc"]]),ie={class:"service-details"},ce={key:3,class:"stack"},oe=L({__name:"ServiceDetailView",props:{page:{},size:{},search:{},query:{},gatewayType:{},mesh:{},service:{}},setup(D){const e=D,g=E(),{t:S}=H(),u=k(null),d=k(null),_=k(!0),p=k(null);b();function b(){C()}async function C(){_.value=!0,p.value=null,u.value=null,d.value=null;const s=e.mesh,c=e.service;try{u.value=await g.getServiceInsight({mesh:s,name:c}),u.value.serviceType==="external"&&(d.value=await g.getExternalServiceByServiceInsightName(s,c))}catch(l){l instanceof Error?p.value=l:console.error(l)}finally{_.value=!1}}return(s,c)=>{const l=A("KCard");return r(),i(U,null,{default:a(({route:m})=>[o(Y,{title:T(S)("services.routes.item.title",{name:m.params.service})},null,8,["title"]),t(),o(M,{breadcrumbs:[{to:{name:"services-list-view",params:{mesh:m.params.mesh}},text:T(S)("services.routes.item.breadcrumbs")}]},{default:a(()=>{var z;return[h("div",ie,[_.value?(r(),i(Q,{key:0})):p.value!==null?(r(),i(ee,{key:1,error:p.value},null,8,["error"])):u.value===null?(r(),i(Z,{key:2})):(r(),$("div",ce,[o(ne,{service:u.value,"external-service":d.value},null,8,["service","external-service"]),t(),((z=u.value)==null?void 0:z.serviceType)!=="external"?(r(),i(X,{key:0,src:`/${m.params.mesh}/dataplanes/for/${m.params.service}/of/${e.gatewayType}?page=${e.page}&size=${e.size}&search=${e.search}`},{default:a(({data:v,error:N})=>{var V;return[(r(!0),$(q,null,K([typeof((V=v==null?void 0:v.items)==null?void 0:V[0].dataplane.networking.gateway)>"u"],B=>(r(),i(l,{key:B},{body:a(()=>[o(te,{"data-testid":"data-plane-collection",class:"data-plane-collection","page-number":e.page,"page-size":e.size,total:v==null?void 0:v.total,items:v==null?void 0:v.items,error:N,gateways:B,onChange:({page:n,size:P})=>{m.update({page:String(n),size:String(P)})}},{toolbar:a(()=>[o(re,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:e.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},zone:{description:"filter by “kuma.io/zone” value"}},onFieldsChange:n=>m.update({query:n.query,s:n.query.length>0?JSON.stringify(n.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"]),t(),B?(r(),i(T(O),{key:0,label:"Type","overlay-label":!0,items:[{label:"All",value:"all"},{label:"Builtin",value:"builtin"},{label:"Delegated",value:"delegated"}].map(n=>({...n,selected:n.value===e.gatewayType})),appearance:"select",onSelected:n=>m.update({gatewayType:String(n.value)})},{"item-template":a(({item:n})=>[t(x(n.label),1)]),_:2},1032,["items","onSelected"])):y("",!0)]),_:2},1032,["page-number","page-size","total","items","error","gateways","onChange"])]),_:2},1024))),128))]}),_:2},1032,["src"])):y("",!0)]))])]}),_:2},1032,["breadcrumbs"])]),_:1})}}});const $e=I(oe,[["__scopeId","data-v-12ceed90"]]);export{$e as default};
