import{Z as D}from"./kongponents.es-7e228e6a.js";import{a as u,D as B}from"./DefinitionListItem-7bd1502c.js";import{_ as C}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-0bfb6b79.js";import{S as P}from"./StatusBadge-174a2d36.js";import{T as V}from"./TagList-7d4ec55b.js";import{T as L}from"./TextWithCopyButton-71e1d5a6.js";import{u as N,_ as A}from"./RouteView.vue_vue_type_script_setup_true_lang-49cffe7a.js";import{d as E,m as i,v as I,o as l,c as o,w as s,e as v,b as t,a as c,t as m,g as d,f as h,F as k,u as O}from"./index-2aa994fe.js";const $={class:"entity-section-list"},j={class:"entity-title"},q=E({__name:"ServiceSummary",props:{service:{type:Object,required:!0},externalService:{type:Object,required:!1,default:null}},setup(g){const e=g,p=N(),b=i(()=>({name:"service-detail-view",params:{service:e.service.name,mesh:e.service.mesh}})),_=i(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.networking.address:e.service.addressPort??null),f=i(()=>{var r;return e.service.serviceType==="external"&&e.externalService!==null?(r=e.externalService.networking.tls)!=null&&r.enabled?"Enabled":"Disabled":null}),x=i(()=>{var r,a;if(e.service.serviceType==="external")return null;{const n=((r=e.service.dataplanes)==null?void 0:r.online)??0,w=((a=e.service.dataplanes)==null?void 0:a.total)??0;return`${n} online / ${w} total`}}),y=i(()=>e.service.serviceType==="external"?null:e.service.status??null),S=i(()=>e.service.serviceType==="external"&&e.externalService!==null?e.externalService.tags:null);async function T(r){if(e.service.serviceType==="external"&&e.externalService!==null){const{mesh:a,name:n}=e.externalService;return await p.getExternalService({mesh:a,name:n},r)}else{const{mesh:a,name:n}=e.service;return await p.getServiceInsight({mesh:a,name:n},r)}}return(r,a)=>{const n=I("router-link");return l(),o(O(D),null,{body:s(()=>[v("div",$,[v("section",null,[v("h1",j,[v("span",null,[t(`
              Service:

              `),c(L,{text:e.service.name},{default:s(()=>[c(n,{to:b.value},{default:s(()=>[t(m(e.service.name),1)]),_:1},8,["to"])]),_:1},8,["text"])]),t(),y.value?(l(),o(P,{key:0,status:y.value},null,8,["status"])):d("",!0)]),t(),c(B,{class:"mt-4"},{default:s(()=>[c(u,{term:"Address"},{default:s(()=>[_.value!==null?(l(),h(k,{key:0},[t(m(_.value),1)],64)):(l(),h(k,{key:1},[t(`
                —
              `)],64))]),_:1}),t(),f.value!==null?(l(),o(u,{key:0,term:"TLS"},{default:s(()=>[t(m(f.value),1)]),_:1})):d("",!0),t(),x.value!==null?(l(),o(u,{key:1,term:"Data Plane Proxies"},{default:s(()=>[t(m(x.value),1)]),_:1})):d("",!0),t(),S.value!==null?(l(),o(u,{key:2,term:"Tags"},{default:s(()=>[c(V,{tags:S.value},null,8,["tags"])]),_:1})):d("",!0)]),_:1})]),t(),c(C,{id:"code-block-service","resource-fetcher":T,"resource-fetcher-watch-key":e.service.name,"is-searchable":"","show-copy-as-kubernetes-button":e.service.serviceType==="external"&&e.externalService!==null,"code-max-height":"250px"},null,8,["resource-fetcher-watch-key","show-copy-as-kubernetes-button"])])]),_:1})}}});const J=A(q,[["__scopeId","data-v-6b783de5"]]);export{J as S};
