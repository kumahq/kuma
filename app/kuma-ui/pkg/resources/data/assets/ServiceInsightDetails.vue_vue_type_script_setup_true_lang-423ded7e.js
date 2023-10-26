import{d as h,g,r as k,o as s,l as d,j as r,w as t,i as c,E as x,x as I,p as B,H as i,k as l,a6 as _,n as a,a1 as y,m as C,F as D,ay as $}from"./index-0d828147.js";import{_ as b}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-25f4cd1f.js";import{T as E}from"./TagList-ababde09.js";import{S as N}from"./StatusBadge-e02331a5.js";const P={key:3,class:"columns"},j=h({__name:"ExternalServiceDetails",props:{mesh:{},service:{}},setup(u){const{t:o}=g(),e=u;return(v,f)=>{const p=k("DataSource");return s(),d("div",null,[r(p,{src:`/meshes/${e.mesh}/external-services/for/${e.service}`},{default:t(({data:n,error:m})=>[m?(s(),c(x,{key:0,error:m},null,8,["error"])):n===void 0?(s(),c(I,{key:1})):n===null?(s(),c(b,{key:2,"data-testid":"no-matching-external-service"},{title:t(()=>[B("p",null,i(l(o)("services.detail.no_matching_external_service",{name:e.service})),1)]),_:1})):(s(),d("div",P,[r(_,null,{title:t(()=>[a(i(l(o)("http.api.property.address")),1)]),body:t(()=>[r(y,{text:n.networking.address},null,8,["text"])]),_:2},1024),a(),n.tags!==null?(s(),c(_,{key:0},{title:t(()=>[a(i(l(o)("http.api.property.tags")),1)]),body:t(()=>[r(E,{tags:n.tags},null,8,["tags"])]),_:2},1024)):C("",!0)]))]),_:1},8,["src"])])}}}),S={class:"columns"},H=h({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(u){const{t:o}=g(),e=u;return(v,f)=>{var p,n;return s(),d("div",S,[r(_,null,{title:t(()=>[a(i(l(o)("http.api.property.status")),1)]),body:t(()=>[r(N,{status:e.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),a(),r(_,null,{title:t(()=>[a(i(l(o)("http.api.property.address")),1)]),body:t(()=>[e.serviceInsight.addressPort?(s(),c(y,{key:0,text:e.serviceInsight.addressPort},null,8,["text"])):(s(),d(D,{key:1},[a(i(l(o)("common.detail.none")),1)],64))]),_:1}),a(),r($,{online:((p=e.serviceInsight.dataplanes)==null?void 0:p.online)??0,total:((n=e.serviceInsight.dataplanes)==null?void 0:n.total)??0},{title:t(()=>[a(i(l(o)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])}}});export{j as _,H as a};
