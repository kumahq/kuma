import{d as v,L as y,o as s,j as m,h as r,w as e,m as g,ae as u,l as a,D as o,i as l,g as i,k,a5 as x,Y as f,ab as S,F as $,ai as b,A as I,p as w,E as B,s as D,_ as V}from"./index-cf0727dc.js";import{T as C}from"./TagList-ae0077fe.js";const T={class:"stack"},E={class:"columns",style:{"--columns":"2"}},N=v({__name:"ExternalServiceDetails",props:{externalService:{}},setup(c){const t=c,{t:n}=y();return(h,_)=>(s(),m("div",T,[r(l(x),null,{body:e(()=>[g("div",E,[r(u,null,{title:e(()=>[a(o(l(n)("http.api.property.address")),1)]),body:e(()=>[a(o(t.externalService.networking.address),1)]),_:1}),a(),t.externalService.tags!==null?(s(),i(u,{key:0},{title:e(()=>[a(o(l(n)("http.api.property.tags")),1)]),body:e(()=>[r(C,{tags:t.externalService.tags},null,8,["tags"])]),_:1})):k("",!0)])]),_:1})]))}}),P={class:"stack"},A={class:"columns",style:{"--columns":"3"}},F=v({__name:"ServiceInsightDetails",props:{serviceInsight:{}},setup(c){const t=c,{t:n}=y();return(h,_)=>(s(),m("div",P,[r(l(x),null,{body:e(()=>{var p,d;return[g("div",A,[r(u,null,{title:e(()=>[a(o(l(n)("http.api.property.status")),1)]),body:e(()=>[r(f,{status:t.serviceInsight.status??"not_available"},null,8,["status"])]),_:1}),a(),r(u,null,{title:e(()=>[a(o(l(n)("http.api.property.address")),1)]),body:e(()=>[t.serviceInsight.addressPort?(s(),i(S,{key:0,text:t.serviceInsight.addressPort},null,8,["text"])):(s(),m($,{key:1},[a(o(l(n)("common.detail.none")),1)],64))]),_:1}),a(),r(b,{online:((p=t.serviceInsight.dataplanes)==null?void 0:p.online)??0,total:((d=t.serviceInsight.dataplanes)==null?void 0:d.total)??0},{title:e(()=>[a(o(l(n)("http.api.property.dataPlaneProxies")),1)]),_:1},8,["online","total"])])]}),_:1})]))}}),J=v({__name:"ServiceDetailView",props:{data:{}},setup(c){const t=c;return(n,h)=>(s(),i(V,{name:"service-detail-view","data-testid":"service-detail-view"},{default:e(({route:_})=>[r(I,null,{default:e(()=>[t.data.serviceType==="external"?(s(),i(w,{key:0,src:`/meshes/${_.params.mesh}/external-services/${_.params.service}`},{default:e(({data:p,error:d})=>[d?(s(),i(B,{key:0,error:d},null,8,["error"])):p===void 0?(s(),i(D,{key:1})):(s(),i(N,{key:2,"external-service":p},null,8,["external-service"]))]),_:2},1032,["src"])):(s(),i(F,{key:1,"service-insight":n.data},null,8,["service-insight"]))]),_:2},1024)]),_:1}))}});export{J as default};
