import{d as f,l as g,a as d,o as t,c as u,e as n,w as e,b as i,p as k,t as c,q as l,a1 as h,f as r,v as w,F as V,W as $}from"./index-784d2bbf.js";import{_ as C}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-f6a2a033.js";import{E as I}from"./ErrorBlock-d38c2168.js";import{_ as b}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-8f5d9bcc.js";import{_ as B}from"./TagList.vue_vue_type_script_setup_true_lang-115cdf7e.js";import{T as x}from"./TextWithCopyButton-7ef74197.js";import{S as D}from"./StatusBadge-a6acfbee.js";import"./index-9dd3e7d3.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-9960c4c9.js";import"./CopyButton-9c00109a.js";const S={key:3,class:"columns"},E=f({__name:"ExternalServiceDetails",props:{mesh:{},service:{}},setup(m){const{t:a}=g(),s=m;return(y,v)=>{const p=d("DataSource");return t(),u("div",null,[n(p,{src:`/meshes/${s.mesh}/external-services/for/${s.service}`},{default:e(({data:o,error:_})=>[_?(t(),i(I,{key:0,error:_},null,8,["error"])):o===void 0?(t(),i(b,{key:1})):o===null?(t(),i(C,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[k("p",null,c(l(a)("services.detail.no_matching_external_service",{name:s.service})),1)]),_:1})):(t(),u("div",S,[n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.address")),1)]),body:e(()=>[n(x,{text:o.networking.address},null,8,["text"])]),_:2},1024),r(),o.tags!==null?(t(),i(h,{key:0},{title:e(()=>[r(c(l(a)("http.api.property.tags")),1)]),body:e(()=>[n(B,{tags:o.tags},null,8,["tags"])]),_:2},1024)):w("",!0)]))]),_:1},8,["src"])])}}}),N={class:"columns"},P=f({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(m){const{t:a}=g(),s=m;return(y,v)=>{var p,o;return t(),u("div",N,[n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.status")),1)]),body:e(()=>[n(D,{status:s.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),r(),n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.address")),1)]),body:e(()=>[s.serviceInsight.addressPort?(t(),i(x,{key:0,text:s.serviceInsight.addressPort},null,8,["text"])):(t(),u(V,{key:1},[r(c(l(a)("common.detail.none")),1)],64))]),_:1}),r(),n($,{online:((p=s.serviceInsight.dataplanes)==null?void 0:p.online)??0,total:((o=s.serviceInsight.dataplanes)==null?void 0:o.total)??0},{title:e(()=>[r(c(l(a)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])}}}),T={class:"stack"},J=f({__name:"ServiceDetailView",props:{data:{}},setup(m){const a=m;return(s,y)=>{const v=d("KCard"),p=d("AppView"),o=d("RouteView");return t(),i(o,{name:"service-detail-view",params:{mesh:"",service:""}},{default:e(({route:_})=>[n(p,null,{default:e(()=>[k("div",T,[n(v,null,{body:e(()=>[a.data.serviceType==="external"?(t(),i(E,{key:0,mesh:_.params.mesh,service:_.params.service},null,8,["mesh","service"])):(t(),i(P,{key:1,"service-insight":s.data},null,8,["service-insight"]))]),_:2},1024)])]),_:2},1024)]),_:1})}}});export{J as default};
