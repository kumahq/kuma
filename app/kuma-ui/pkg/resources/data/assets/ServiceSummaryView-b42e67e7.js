import{d as A,l as I,u as D,a as v,o as t,b as l,w as e,e as r,p as m,f as s,q as a,t as i,c as d,a1 as y,s as x,F as N,W as P,z as q,A as E,a4 as L,_ as F}from"./index-c2f88a6f.js";import{_ as f}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-89b4cabc.js";import{E as S}from"./ErrorBlock-ea8aec14.js";import{_ as V}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-a42aed1c.js";import{_ as Q}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-55c9b654.js";import{S as W}from"./StatusBadge-4ecc75dc.js";import{T as z}from"./TagList-60c8b15c.js";import{T}from"./TextWithCopyButton-5e8334c0.js";import"./index-52545d1d.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-5a42a90e.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-c936cb45.js";import"./CopyButton-989013bb.js";import"./toYaml-4e00099e.js";const K=_=>(q("data-v-07406fdd"),_=_(),E(),_),j={class:"summary-title-wrapper"},G=K(()=>m("img",{"aria-hidden":"true",src:L},null,-1)),H={class:"summary-title"},J={key:1,class:"stack"},M={class:"mt-4"},O={key:0},U={key:3,class:"stack"},X={key:1,class:"stack"},Y={key:0},Z={class:"mt-4"},ee=A({__name:"ServiceSummaryView",props:{service:{default:void 0}},setup(_){const{t:o}=I(),b=D(),c=_;return(te,se)=>{const C=v("RouteTitle"),B=v("RouterLink"),k=v("DataSource"),R=v("AppView"),$=v("RouteView");return t(),l($,{name:"service-summary-view",params:{mesh:"",service:"",codeSearch:""}},{default:e(({route:p})=>[r(R,null,{title:e(()=>[m("div",j,[G,s(),m("h2",H,[r(B,{to:{name:"service-detail-view",params:{service:p.params.service}}},{default:e(()=>[r(C,{title:a(o)("services.routes.item.title",{name:p.params.service})},null,8,["title"])]),_:2},1032,["to"])])])]),default:e(()=>{var g,w;return[s(),c.service===void 0?(t(),l(f,{key:0},{message:e(()=>[m("p",null,i(a(o)("common.collection.summary.empty_message",{type:"Service"})),1)]),default:e(()=>[s(i(a(o)("common.collection.summary.empty_title",{type:"Service"}))+" ",1)]),_:1})):(t(),d("div",J,[m("div",null,[m("h3",null,i(a(o)("services.routes.item.overview")),1),s(),m("div",M,[c.service.serviceType==="external"?(t(),d("div",O,[r(k,{src:`/meshes/${c.service.mesh}/external-services/for/${c.service.name}`},{default:e(({data:n,error:u})=>[u?(t(),l(S,{key:0,error:u},null,8,["error"])):n===void 0?(t(),l(V,{key:1})):n===null?(t(),l(f,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[m("p",null,i(a(o)("services.detail.no_matching_external_service",{name:c.service.name})),1)]),_:1})):(t(),d("div",U,[r(y,null,{title:e(()=>[s(i(a(o)("http.api.property.address")),1)]),body:e(()=>[r(T,{text:n.networking.address},null,8,["text"])]),_:2},1024),s(),n.tags!==null?(t(),l(y,{key:0},{title:e(()=>[s(i(a(o)("http.api.property.tags")),1)]),body:e(()=>[r(z,{tags:n.tags},null,8,["tags"])]),_:2},1024)):x("",!0)]))]),_:1},8,["src"])])):(t(),d("div",X,[r(y,null,{title:e(()=>[s(i(a(o)("http.api.property.status")),1)]),body:e(()=>[r(W,{status:c.service.status??"not_available"},null,8,["status"])]),_:1}),s(),r(y,null,{title:e(()=>[s(i(a(o)("http.api.property.address")),1)]),body:e(()=>[c.service.addressPort?(t(),l(T,{key:0,text:c.service.addressPort},null,8,["text"])):(t(),d(N,{key:1},[s(i(a(o)("common.detail.none")),1)],64))]),_:1}),s(),r(P,{online:((g=c.service.dataplanes)==null?void 0:g.online)??0,total:((w=c.service.dataplanes)==null?void 0:w.total)??0},{title:e(()=>[s(i(a(o)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])]))])]),s(),c.service.serviceType==="external"?(t(),d("div",Y,[m("h3",null,i(a(o)("services.routes.item.config")),1),s(),m("div",Z,[r(k,{src:`/meshes/${p.params.mesh}/external-services/for/${p.params.service}`},{default:e(({data:n,error:u})=>[u?(t(),l(S,{key:0,error:u},null,8,["error"])):n===void 0?(t(),l(V,{key:1})):n===null?(t(),l(f,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[m("p",null,i(a(o)("services.detail.no_matching_external_service",{name:p.params.service})),1)]),_:2},1024)):(t(),l(Q,{key:3,id:"code-block-service",resource:n,"resource-fetcher":h=>a(b).getExternalService({mesh:n.mesh,name:n.name},h),"is-searchable":"",query:p.params.codeSearch,onQueryChange:h=>p.update({codeSearch:h})},null,8,["resource","resource-fetcher","query","onQueryChange"]))]),_:2},1032,["src"])])])):x("",!0)]))]}),_:2},1024)]),_:1})}}});const ye=F(ee,[["__scopeId","data-v-07406fdd"]]);export{ye as default};
