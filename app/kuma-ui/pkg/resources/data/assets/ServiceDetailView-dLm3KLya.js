import{d as K,a as y,o as t,b as r,w as s,E as k,A as b,a2 as V,t as l,m as g,Z as v,f as n,W as R,e as i,p as u,c,F as f,av as A,T as M,K as E,q as w,U as W,D as Z,_ as O}from"./index-1j9z4Egf.js";import{A as Q}from"./AppCollection-QEcgTHkF.js";import{_ as U}from"./ResourceCodeBlock.vue_vue_type_style_index_0_lang-owWKimIX.js";import{F as J}from"./FilterBar-ZHB4gAdZ.js";import{S as F}from"./StatusBadge-j_uZdXEO.js";import{S as j}from"./SummaryView-GaIQsIB6.js";import{T as G}from"./TagList-X1LjfAG5.js";import"./CodeBlock-ejECcgv-.js";import"./toYaml-sPaYOD3i.js";const H={key:2,class:"stack"},X={key:0},Y={key:3,class:"columns"},D={key:1,class:"columns"},ee={key:0},ae={key:0},se={key:1},te={key:1},ne={class:"mt-4"},ie=K({__name:"ServiceDetailView",setup(oe){return(re,le)=>{const h=y("DataSource"),$=y("KCard"),x=y("RouterLink"),S=y("KTooltip"),B=y("RouterView"),I=y("AppView"),L=y("RouteView");return t(),r(h,{src:"/me"},{default:s(({data:q})=>[q?(t(),r(L,{key:0,name:"service-detail-view",params:{mesh:"",service:"",page:1,size:q.pageSize,query:"",s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:s(({can:C,route:a,t:o})=>[i(I,null,{default:s(()=>[i(h,{src:`/meshes/${a.params.mesh}/service-insights/${a.params.service}`},{default:s(({data:d,error:T})=>[T?(t(),r(k,{key:0,error:T},null,8,["error"])):d===void 0?(t(),r(b,{key:1})):(t(),c("div",H,[i($,null,{default:s(()=>{var m,_;return[!C("use gateways ui")&&d.serviceType==="external"?(t(),c("div",X,[i(h,{src:`/meshes/${a.params.mesh}/external-services/for/${a.params.service}`},{default:s(({data:e,error:p})=>[p?(t(),r(k,{key:0,error:p},null,8,["error"])):e===void 0?(t(),r(b,{key:1})):e===null?(t(),r(V,{key:2,"data-testid":"no-matching-external-service"},{title:s(()=>[g("p",null,l(o("services.detail.no_matching_external_service",{name:a.params.service})),1)]),_:2},1024)):(t(),c("div",Y,[i(v,null,{title:s(()=>[n(l(o("http.api.property.address")),1)]),body:s(()=>[i(R,{text:e.networking.address},null,8,["text"])]),_:2},1024),n(),e.tags!==null?(t(),r(v,{key:0},{title:s(()=>[n(l(o("http.api.property.tags")),1)]),body:s(()=>[i(G,{tags:e.tags},null,8,["tags"])]),_:2},1024)):u("",!0)]))]),_:2},1032,["src"])])):(t(),c("div",D,[i(v,null,{title:s(()=>[n(l(o("http.api.property.status")),1)]),body:s(()=>[i(F,{status:d.status},null,8,["status"])]),_:2},1024),n(),i(v,null,{title:s(()=>[n(l(o("http.api.property.address")),1)]),body:s(()=>[d.addressPort?(t(),r(R,{key:0,text:d.addressPort},null,8,["text"])):(t(),c(f,{key:1},[n(l(o("common.detail.none")),1)],64))]),_:2},1024),n(),i(A,{online:((m=d.dataplanes)==null?void 0:m.online)??0,total:((_=d.dataplanes)==null?void 0:_.total)??0},{title:s(()=>[n(l(o("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])]))]}),_:2},1024),n(),d.serviceType!=="external"?(t(),c("div",ee,[g("h3",null,l(o("services.detail.data_plane_proxies")),1),n(),i($,{class:"mt-4"},{default:s(()=>[i(h,{src:`/meshes/${a.params.mesh}/dataplanes/for/${a.params.service}?page=${a.params.page}&size=${a.params.size}&search=${a.params.s}`},{default:s(({data:m,error:_})=>[_!==void 0?(t(),r(k,{key:0,error:_},null,8,["error"])):(t(),r(Q,{key:1,class:"data-plane-collection","data-testid":"data-plane-collection","page-number":a.params.page,"page-size":a.params.size,headers:[{label:"Name",key:"name"},...C("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:m==null?void 0:m.items,total:m==null?void 0:m.total,error:_,"is-selected-row":e=>e.name===a.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":o("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":o("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":o("common.documentation"),onChange:a.update},{toolbar:s(()=>[i(J,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:a.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...C("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onFieldsChange:e=>a.update({query:e.query,s:e.query.length>0?JSON.stringify(e.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"])]),name:s(({row:e})=>[i(x,{class:"name-link",title:e.name,to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.name},query:{page:a.params.page,size:a.params.size,query:a.params.query}}},{default:s(()=>[n(l(e.name),1)]),_:2},1032,["title","to"])]),zone:s(({row:e})=>[e.zone?(t(),r(x,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[n(l(e.zone),1)]),_:2},1032,["to"])):(t(),c(f,{key:1},[n(l(o("common.collection.none")),1)],64))]),certificate:s(({row:e})=>{var p;return[(p=e.dataplaneInsight.mTLS)!=null&&p.certificateExpirationTime?(t(),c(f,{key:0},[n(l(o("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(t(),c(f,{key:1},[n(l(o("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:s(({row:e})=>[i(F,{status:e.status},null,8,["status"])]),warnings:s(({row:e})=>[e.isCertExpired||e.warnings.length>0?(t(),r(S,{key:0},{content:s(()=>[g("ul",null,[e.warnings.length>0?(t(),c("li",ae,l(o("data-planes.components.data-plane-list.version_mismatch")),1)):u("",!0),n(),e.isCertExpired?(t(),c("li",se,l(o("data-planes.components.data-plane-list.cert_expired")),1)):u("",!0)])]),default:s(()=>[n(),i(M,{class:"mr-1",size:w(E),"hide-title":""},null,8,["size"])]),_:2},1024)):(t(),c(f,{key:1},[n(l(o("common.collection.none")),1)],64))]),details:s(({row:e})=>[i(x,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.name}}},{default:s(()=>[n(l(o("common.collection.details_link"))+" ",1),i(w(W),{display:"inline-block",decorative:"",size:w(E)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","error","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])),n(),a.params.dataPlane?(t(),r(B,{key:2},{default:s(e=>[i(j,{onClose:p=>a.replace({name:a.name,params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size}})},{default:s(()=>[(t(),r(Z(e.Component),{name:a.params.dataPlane,"dataplane-overview":m==null?void 0:m.items.find(p=>p.name===a.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):u("",!0)]),_:2},1032,["src"])]),_:2},1024)])):u("",!0),n(),d.serviceType==="external"?(t(),c("div",te,[g("h3",null,l(o("services.detail.config")),1),n(),g("div",ne,[i(h,{src:`/meshes/${a.params.mesh}/external-services/for/${a.params.service}`},{default:s(({data:m,error:_})=>[_?(t(),r(k,{key:0,error:_},null,8,["error"])):m===void 0?(t(),r(b,{key:1})):m===null?(t(),r(V,{key:2,"data-testid":"no-matching-external-service"},{title:s(()=>[g("p",null,l(o("services.detail.no_matching_external_service",{name:a.params.service})),1)]),_:2},1024)):(t(),r(U,{key:3,"data-testid":"external-service-config",resource:m.config,"is-searchable":"",query:a.params.codeSearch,"is-filter-mode":a.params.codeFilter,"is-reg-exp-mode":a.params.codeRegExp,onQueryChange:e=>a.update({codeSearch:e}),onFilterModeChange:e=>a.update({codeFilter:e}),onRegExpModeChange:e=>a.update({codeRegExp:e})},{default:s(({copy:e,copying:p})=>[p?(t(),r(h,{key:0,src:`/meshes/${m.mesh}/external-service/${m.name}/as/kubernetes?no-store`,onChange:z=>{e(P=>P(z))},onError:z=>{e((P,N)=>N(z))}},null,8,["src","onChange","onError"])):u("",!0)]),_:2},1032,["resource","query","is-filter-mode","is-reg-exp-mode","onQueryChange","onFilterModeChange","onRegExpModeChange"]))]),_:2},1032,["src"])])])):u("",!0)]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):u("",!0)]),_:1})}}}),fe=O(ie,[["__scopeId","data-v-0e99e371"]]);export{fe as default};
