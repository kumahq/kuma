import{d as y,o as r,e as h,h as a,w as e,i as g,g as s,t as p,b as i,a as c,f as S,l as x,F as b}from"./index-f1b8ae6a.js";import{l as $,h as f,D as _,T as k,S as I,R as w,A as B,i as D,n as T,E as V,_ as C}from"./RouteView.vue_vue_type_script_setup_true_lang-4a32e1ca.js";import{_ as E}from"./ResourceCodeBlock.vue_vue_type_script_setup_true_lang-be27b925.js";import{T as A}from"./TagList-386ee57e.js";import{_ as N}from"./RouteTitle.vue_vue_type_script_setup_true_lang-6484968f.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-14dd845b.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-d1d1c408.js";import"./toYaml-4e00099e.js";const P={class:"stack"},F={class:"columns",style:{"--columns":"3"}},R=y({__name:"ExternalServiceDetails",props:{serviceInsight:{},externalService:{}},setup(u){const t=u,n=$(),{t:d}=f();async function v(l){const{mesh:o,name:m}=t.externalService;return await n.getExternalService({mesh:o,name:m},l)}return(l,o)=>(r(),h("div",P,[a(i(x),null,{body:e(()=>[g("div",F,[a(_,null,{title:e(()=>[s(p(i(d)("http.api.property.name")),1)]),body:e(()=>[a(k,{text:t.serviceInsight.name},null,8,["text"])]),_:1}),s(),a(_,null,{title:e(()=>[s(p(i(d)("http.api.property.address")),1)]),body:e(()=>[s(p(t.externalService.networking.address),1)]),_:1}),s(),t.externalService.tags!==null?(r(),c(_,{key:0},{title:e(()=>[s(p(i(d)("http.api.property.tags")),1)]),body:e(()=>[a(A,{tags:t.externalService.tags},null,8,["tags"])]),_:1})):S("",!0)])]),_:1}),s(),a(E,{id:"code-block-service",resource:t.externalService,"resource-fetcher":v,"is-searchable":""},null,8,["resource"])]))}}),K={class:"stack"},L={class:"columns",style:{"--columns":"3"}},W=y({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(u){const t=u,{t:n}=f();return(d,v)=>(r(),h("div",K,[a(i(x),null,{body:e(()=>{var l,o;return[g("div",L,[a(_,null,{title:e(()=>[s(p(i(n)("http.api.property.status")),1)]),body:e(()=>[a(I,{status:t.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),s(),a(_,null,{title:e(()=>[s(p(i(n)("http.api.property.address")),1)]),body:e(()=>[t.serviceInsight.addressPort?(r(),c(k,{key:0,text:t.serviceInsight.addressPort},null,8,["text"])):(r(),h(b,{key:1},[s(p(i(n)("common.detail.none")),1)],64))]),_:1}),s(),a(w,{online:((l=t.serviceInsight.dataplanes)==null?void 0:l.online)??0,total:((o=t.serviceInsight.dataplanes)==null?void 0:o.total)??0},{title:e(()=>[s(p(i(n)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])]}),_:1})]))}}),Q=y({__name:"ServiceDetailView",props:{data:{}},setup(u){const t=u,{t:n}=f();return(d,v)=>(r(),c(C,{name:"service-detail-view","data-testid":"service-detail-view"},{default:e(({route:l})=>[a(B,null,{title:e(()=>[g("h2",null,[a(N,{title:i(n)("services.routes.item.navigation.service-detail-view"),render:!0},null,8,["title"])])]),default:e(()=>[s(),t.data.serviceType==="external"?(r(),c(D,{key:0,src:`/meshes/${l.params.mesh}/external-services/${l.params.service}`},{default:e(({data:o,error:m})=>[o===void 0?(r(),c(T,{key:0})):m?(r(),c(V,{key:1,error:m},null,8,["error"])):(r(),c(R,{key:2,"service-insight":d.data,"external-service":o},null,8,["service-insight","external-service"]))]),_:2},1032,["src"])):(r(),c(W,{key:1,"service-insight":d.data},null,8,["service-insight"]))]),_:2},1024)]),_:1}))}});export{Q as default};
