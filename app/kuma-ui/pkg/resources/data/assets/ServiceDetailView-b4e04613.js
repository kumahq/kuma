import{d as f,l as y,a as d,o as t,c as u,e as n,w as e,b as i,p as k,t as c,q as l,a0 as h,f as r,s as V,F as w,V as C}from"./index-e9fbefd3.js";import{_ as I}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-15e0e5b5.js";import{E as $}from"./ErrorBlock-a3710a04.js";import{_ as B}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-5b5238ed.js";import{T as D}from"./TagList-e390efe3.js";import{T as x}from"./TextWithCopyButton-0bfc7306.js";import{S as b}from"./StatusBadge-494c559b.js";import"./index-fce48c05.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-060e8475.js";import"./CopyButton-6f1494f2.js";const T={key:3,class:"columns"},S=f({__name:"ExternalServiceDetails",props:{mesh:{},service:{}},setup(m){const{t:a}=y(),s=m;return(g,v)=>{const p=d("DataSource");return t(),u("div",null,[n(p,{src:`/meshes/${s.mesh}/external-services/for/${s.service}`},{default:e(({data:o,error:_})=>[_?(t(),i($,{key:0,error:_},null,8,["error"])):o===void 0?(t(),i(B,{key:1})):o===null?(t(),i(I,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[k("p",null,c(l(a)("services.detail.no_matching_external_service",{name:s.service})),1)]),_:1})):(t(),u("div",T,[n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.address")),1)]),body:e(()=>[n(x,{text:o.networking.address},null,8,["text"])]),_:2},1024),r(),o.tags!==null?(t(),i(h,{key:0},{title:e(()=>[r(c(l(a)("http.api.property.tags")),1)]),body:e(()=>[n(D,{tags:o.tags},null,8,["tags"])]),_:2},1024)):V("",!0)]))]),_:1},8,["src"])])}}}),E={class:"columns"},N=f({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(m){const{t:a}=y(),s=m;return(g,v)=>{var p,o;return t(),u("div",E,[n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.status")),1)]),body:e(()=>[n(b,{status:s.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),r(),n(h,null,{title:e(()=>[r(c(l(a)("http.api.property.address")),1)]),body:e(()=>[s.serviceInsight.addressPort?(t(),i(x,{key:0,text:s.serviceInsight.addressPort},null,8,["text"])):(t(),u(w,{key:1},[r(c(l(a)("common.detail.none")),1)],64))]),_:1}),r(),n(C,{online:((p=s.serviceInsight.dataplanes)==null?void 0:p.online)??0,total:((o=s.serviceInsight.dataplanes)==null?void 0:o.total)??0},{title:e(()=>[r(c(l(a)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])}}}),P={class:"stack"},H=f({__name:"ServiceDetailView",props:{data:{}},setup(m){const a=m;return(s,g)=>{const v=d("KCard"),p=d("AppView"),o=d("RouteView");return t(),i(o,{name:"service-detail-view",params:{mesh:"",service:""}},{default:e(({route:_})=>[n(p,null,{default:e(()=>[k("div",P,[n(v,null,{default:e(()=>[a.data.serviceType==="external"?(t(),i(S,{key:0,mesh:_.params.mesh,service:_.params.service},null,8,["mesh","service"])):(t(),i(N,{key:1,"service-insight":s.data},null,8,["service-insight"]))]),_:2},1024)])]),_:2},1024)]),_:1})}}});export{H as default};
