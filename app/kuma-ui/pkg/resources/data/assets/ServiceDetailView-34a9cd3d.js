import{d as g,g as S,o as t,l as y,j as i,w as e,p as x,a9 as d,n as a,H as c,k as o,i as r,m as $,$ as w,W as I,a5 as b,F as B,ae as D,r as h,E as C,x as T,af as E}from"./index-df368167.js";import{T as N}from"./TagList-cb56e619.js";const P={class:"stack"},R={class:"columns"},A=g({__name:"ExternalServiceDetails",props:{externalService:{}},setup(l){const s=l,{t:n}=S();return(k,u)=>(t(),y("div",P,[i(o(w),null,{body:e(()=>[x("div",R,[i(d,null,{title:e(()=>[a(c(o(n)("http.api.property.address")),1)]),body:e(()=>[a(c(s.externalService.networking.address),1)]),_:1}),a(),s.externalService.tags!==null?(t(),r(d,{key:0},{title:e(()=>[a(c(o(n)("http.api.property.tags")),1)]),body:e(()=>[i(N,{tags:s.externalService.tags},null,8,["tags"])]),_:1})):$("",!0)])]),_:1})]))}}),F={class:"stack"},W={class:"columns"},j=g({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(l){const s=l,{t:n}=S();return(k,u)=>(t(),y("div",F,[i(o(w),null,{body:e(()=>{var p,_;return[x("div",W,[i(d,null,{title:e(()=>[a(c(o(n)("http.api.property.status")),1)]),body:e(()=>[i(I,{status:s.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),a(),i(d,null,{title:e(()=>[a(c(o(n)("http.api.property.address")),1)]),body:e(()=>[s.serviceInsight.addressPort?(t(),r(b,{key:0,text:s.serviceInsight.addressPort},null,8,["text"])):(t(),y(B,{key:1},[a(c(o(n)("common.detail.none")),1)],64))]),_:1}),a(),i(D,{online:((p=s.serviceInsight.dataplanes)==null?void 0:p.online)??0,total:((_=s.serviceInsight.dataplanes)==null?void 0:_.total)??0},{title:e(()=>[a(c(o(n)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])]}),_:1})]))}}),q=g({__name:"ServiceDetailView",props:{data:{}},setup(l){const s=l;return(n,k)=>{const u=h("DataSource"),p=h("AppView"),_=h("RouteView");return t(),r(_,{name:"service-detail-view",params:{mesh:"",service:""}},{default:e(({route:m,t:V})=>[i(p,null,{default:e(()=>[s.data.serviceType==="external"?(t(),r(u,{key:0,src:`/meshes/${m.params.mesh}/external-services/for/${m.params.service}`},{default:e(({data:v,error:f})=>[f?(t(),r(C,{key:0,error:f},null,8,["error"])):v===void 0?(t(),r(T,{key:1})):v===null?(t(),r(E,{key:2,"data-testid":"no-matching-external-service"},{title:e(()=>[x("p",null,c(V("services.detail.no_matching_external_service",{name:m.params.service})),1)]),_:2},1024)):(t(),r(A,{key:3,"external-service":v},null,8,["external-service"]))]),_:2},1032,["src"])):(t(),r(j,{key:1,"service-insight":n.data},null,8,["service-insight"]))]),_:2},1024)]),_:1})}}});export{q as default};
