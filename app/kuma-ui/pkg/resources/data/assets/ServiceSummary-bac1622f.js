import{z as D}from"./kongponents.es-176426e1.js";import{a as u,D as B}from"./DefinitionListItem-56ada3b3.js";import{_ as C}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-d548267a.js";import{S as P}from"./StatusBadge-4b1e7b88.js";import{T as V}from"./TagList-b39ced52.js";import{T as L}from"./TextWithCopyButton-40f68948.js";import{u as N}from"./index-4917657d.js";import{d as j,c as i,x as A,o as n,b as o,w as l,i as v,h as t,g as c,f as m,j as h,t as d,F as g,e as E}from"./index-8d80a271.js";import{_ as I}from"./_plugin-vue_export-helper-c27b6911.js";const O={class:"entity-section-list"},$={class:"entity-title"},q=j({__name:"ServiceSummary",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null}},setup(k){const e=k,f=N(),T=i(()=>({name:"service-detail-view",params:{service:e.service.name,mesh:e.service.mesh}})),p=i(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.networking.address:e.service.addressPort??null),_=i(()=>{var r;return e.service.serviceType==="external"&&e.externalService!==null?(r=e.externalService.networking.tls)!=null&&r.enabled?"Enabled":"Disabled":null}),x=i(()=>{var r,s;if(e.service.serviceType==="external")return null;{const a=((r=e.service.dataplanes)==null?void 0:r.online)??0,w=((s=e.service.dataplanes)==null?void 0:s.total)??0;return`${a} online / ${w} total`}}),y=i(()=>e.service.serviceType==="external"?null:e.service.status??null),S=i(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.tags:null);async function b(r){if(e.service.serviceType==="external"&&e.externalService!==null){const{mesh:s,name:a}=e.externalService;return await f.getExternalService({mesh:s,name:a},r)}else{const{mesh:s,name:a}=e.service;return await f.getServiceInsight({mesh:s,name:a},r)}}return(r,s)=>{const a=A("router-link");return n(),o(E(D),null,{body:l(()=>[v("div",O,[v("section",null,[v("h1",$,[v("span",null,[t(`
              Service:

              `),c(a,{to:T.value},{default:l(()=>[c(L,{text:e.service.name},null,8,["text"])]),_:1},8,["to"])]),t(),y.value?(n(),o(P,{key:0,status:y.value},null,8,["status"])):m("",!0)]),t(),c(B,{class:"mt-4"},{default:l(()=>[c(u,{term:"Address"},{default:l(()=>[p.value!==null?(n(),h(g,{key:0},[t(d(p.value),1)],64)):(n(),h(g,{key:1},[t(`
                —
              `)],64))]),_:1}),t(),_.value!==null?(n(),o(u,{key:0,term:"TLS"},{default:l(()=>[t(d(_.value),1)]),_:1})):m("",!0),t(),x.value!==null?(n(),o(u,{key:1,term:"Data Plane Proxies"},{default:l(()=>[t(d(x.value),1)]),_:1})):m("",!0),t(),S.value!==null?(n(),o(u,{key:2,term:"Tags"},{default:l(()=>[c(V,{tags:S.value},null,8,["tags"])]),_:1})):m("",!0)]),_:1})]),t(),c(C,{id:"code-block-service","resource-fetcher":b,"resource-fetcher-watch-key":e.service.name,"is-searchable":"","code-max-height":"250px"},null,8,["resource-fetcher-watch-key"])])]),_:1})}}});const Q=I(q,[["__scopeId","data-v-4ffb729e"]]);export{Q as S};
