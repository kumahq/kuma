import{d as B,i as p,o as t,a as m,w as a,j as i,a1 as x,a9 as E,b as c,g as f,V as C,k as n,t as l,a3 as N,H as u,W as L,e as y,A as S,K as A,r as T,_ as q}from"./index-CyAtMQ3G.js";import{p as F}from"./kong-icons.es249-vgKX97Et.js";import{A as K}from"./AppCollection-DSdmYcz_.js";import{F as j}from"./FilterBar-BoTvl30_.js";import{S as V}from"./StatusBadge-yau4rj5h.js";import{S as O}from"./SummaryView-Bn2oVwri.js";import"./kong-icons.es245-BjB891cP.js";const W={key:2,class:"stack"},X={class:"columns"},Z={key:0},H={key:1},U=B({__name:"ServiceDetailView",setup(G){return(J,M)=>{const v=p("KCard"),g=p("RouterLink"),P=p("XIcon"),$=p("RouterView"),k=p("DataSource"),I=p("AppView"),R=p("RouteView");return t(),m(k,{src:"/me"},{default:a(({data:z})=>[z?(t(),m(R,{key:0,name:"service-detail-view",params:{mesh:"",service:"",page:1,size:z.pageSize,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({can:b,route:s,t:o})=>[i(I,null,{default:a(()=>[i(k,{src:`/meshes/${s.params.mesh}/service-insights/${s.params.service}`},{default:a(({data:d,error:w})=>[w?(t(),m(x,{key:0,error:w},null,8,["error"])):d===void 0?(t(),m(E,{key:1})):(t(),c("div",W,[i(v,null,{default:a(()=>{var r,_;return[f("div",X,[i(C,null,{title:a(()=>[n(l(o("http.api.property.status")),1)]),body:a(()=>[i(V,{status:d.status},null,8,["status"])]),_:2},1024),n(),i(C,null,{title:a(()=>[n(l(o("http.api.property.address")),1)]),body:a(()=>[d.addressPort?(t(),m(N,{key:0,text:d.addressPort},null,8,["text"])):(t(),c(u,{key:1},[n(l(o("common.detail.none")),1)],64))]),_:2},1024),n(),i(L,{online:((r=d.dataplanes)==null?void 0:r.online)??0,total:((_=d.dataplanes)==null?void 0:_.total)??0},{title:a(()=>[n(l(o("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1024),n(),f("div",null,[f("h3",null,l(o("services.detail.data_plane_proxies")),1),n(),i(v,{class:"mt-4"},{default:a(()=>[i(k,{src:`/meshes/${s.params.mesh}/dataplanes/for/service-insight/${s.params.service}?page=${s.params.page}&size=${s.params.size}&search=${s.params.s}`},{default:a(({data:r,error:_})=>[_!==void 0?(t(),m(x,{key:0,error:_},null,8,["error"])):(t(),m(K,{key:1,class:"data-plane-collection","data-testid":"data-plane-collection","page-number":s.params.page,"page-size":s.params.size,headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...b("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:r==null?void 0:r.items,total:r==null?void 0:r.total,error:_,"is-selected-row":e=>e.name===s.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":o("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":o("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":o("common.documentation"),onChange:s.update},{toolbar:a(()=>[i(j,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...b("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>s.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"])]),name:a(({row:e})=>[i(g,{class:"name-link",to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:a(()=>[n(l(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(l(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(t(),m(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(l(e.zone),1)]),_:2},1032,["to"])):(t(),c(u,{key:1},[n(l(o("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var h;return[(h=e.dataplaneInsight.mTLS)!=null&&h.certificateExpirationTime?(t(),c(u,{key:0},[n(l(o("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(t(),c(u,{key:1},[n(l(o("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[i(V,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(t(),m(P,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[f("ul",null,[e.warnings.length>0?(t(),c("li",Z,l(o("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),n(),e.isCertExpired?(t(),c("li",H,l(o("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(t(),c(u,{key:1},[n(l(o("common.collection.none")),1)],64))]),details:a(({row:e})=>[i(g,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[n(l(o("common.collection.details_link"))+" ",1),i(S(F),{decorative:"",size:S(A)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","error","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])),n(),s.params.dataPlane?(t(),m($,{key:2},{default:a(e=>[i(O,{onClose:h=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:a(()=>[typeof r<"u"?(t(),m(T(e.Component),{key:0,items:r.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):y("",!0)]),_:2},1032,["src"])]),_:2},1024)])]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):y("",!0)]),_:1})}}}),ne=q(U,[["__scopeId","data-v-78159475"]]);export{ne as default};
