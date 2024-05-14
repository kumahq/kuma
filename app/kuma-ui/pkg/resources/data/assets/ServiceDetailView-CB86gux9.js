import{d as E,a as p,o as t,b as m,w as a,e as i,a0 as b,ad as R,c,m as f,a7 as w,f as n,t as l,T as N,F as _,aJ as T,q as y,p as S,Q as L,K as q,E as F,_ as A}from"./index-T0BkiAMa.js";import{A as K}from"./AppCollection--hpbcKQi.js";import{F as O}from"./FilterBar-BryPXAa3.js";import{S as V}from"./StatusBadge-HOzsSfnt.js";import{S as W}from"./SummaryView-DIa7F5ow.js";const X={key:2,class:"stack"},Z={class:"columns"},j={key:0},J={key:1},Q=E({__name:"ServiceDetailView",setup(U){return(G,H)=>{const v=p("KCard"),g=p("RouterLink"),P=p("XIcon"),I=p("RouterView"),h=p("DataSource"),$=p("AppView"),B=p("RouteView");return t(),m(h,{src:"/me"},{default:a(({data:z})=>[z?(t(),m(B,{key:0,name:"service-detail-view",params:{mesh:"",service:"",page:1,size:z.pageSize,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({can:x,route:s,t:o})=>[i($,null,{default:a(()=>[i(h,{src:`/meshes/${s.params.mesh}/service-insights/${s.params.service}`},{default:a(({data:d,error:C})=>[C?(t(),m(b,{key:0,error:C},null,8,["error"])):d===void 0?(t(),m(R,{key:1})):(t(),c("div",X,[i(v,null,{default:a(()=>{var r,u;return[f("div",Z,[i(w,null,{title:a(()=>[n(l(o("http.api.property.status")),1)]),body:a(()=>[i(V,{status:d.status},null,8,["status"])]),_:2},1024),n(),i(w,null,{title:a(()=>[n(l(o("http.api.property.address")),1)]),body:a(()=>[d.addressPort?(t(),m(N,{key:0,text:d.addressPort},null,8,["text"])):(t(),c(_,{key:1},[n(l(o("common.detail.none")),1)],64))]),_:2},1024),n(),i(T,{online:((r=d.dataplanes)==null?void 0:r.online)??0,total:((u=d.dataplanes)==null?void 0:u.total)??0},{title:a(()=>[n(l(o("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1024),n(),f("div",null,[f("h3",null,l(o("services.detail.data_plane_proxies")),1),n(),i(v,{class:"mt-4"},{default:a(()=>[i(h,{src:`/meshes/${s.params.mesh}/dataplanes/for/${s.params.service}?page=${s.params.page}&size=${s.params.size}&search=${s.params.s}`},{default:a(({data:r,error:u})=>[u!==void 0?(t(),m(b,{key:0,error:u},null,8,["error"])):(t(),m(K,{key:1,class:"data-plane-collection","data-testid":"data-plane-collection","page-number":s.params.page,"page-size":s.params.size,headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...x("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:r==null?void 0:r.items,total:r==null?void 0:r.total,error:u,"is-selected-row":e=>e.name===s.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":o("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":o("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":o("common.documentation"),onChange:s.update},{toolbar:a(()=>[i(O,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...x("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>s.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"])]),name:a(({row:e})=>[i(g,{class:"name-link",to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:a(()=>[n(l(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(l(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(t(),m(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(l(e.zone),1)]),_:2},1032,["to"])):(t(),c(_,{key:1},[n(l(o("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var k;return[(k=e.dataplaneInsight.mTLS)!=null&&k.certificateExpirationTime?(t(),c(_,{key:0},[n(l(o("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(t(),c(_,{key:1},[n(l(o("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[i(V,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(t(),m(P,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[f("ul",null,[e.warnings.length>0?(t(),c("li",j,l(o("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),n(),e.isCertExpired?(t(),c("li",J,l(o("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(t(),c(_,{key:1},[n(l(o("common.collection.none")),1)],64))]),details:a(({row:e})=>[i(g,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[n(l(o("common.collection.details_link"))+" ",1),i(S(L),{decorative:"",size:S(q)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","error","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])),n(),s.params.dataPlane?(t(),m(I,{key:2},{default:a(e=>[i(W,{onClose:k=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:a(()=>[typeof r<"u"?(t(),m(F(e.Component),{key:0,items:r.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):y("",!0)]),_:2},1032,["src"])]),_:2},1024)])]))]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):y("",!0)]),_:1})}}}),se=A(Q,[["__scopeId","data-v-3078ae0a"]]);export{se as default};
