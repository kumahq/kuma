import{d as $,a as r,o as l,b as p,w as a,e as n,m as u,W as C,f as s,t as m,T as L,c,F as y,ax as R,q as _,R as B,p as g,K as v,S as E,C as I,_ as N}from"./index-C1qiy_FS.js";import{A as K}from"./AppCollection-DnxF72zZ.js";import{F as q}from"./FilterBar-CryFINzN.js";import{S as x}from"./StatusBadge-VZcN4L_G.js";import{S as A}from"./SummaryView-COWmU_Ly.js";const F={class:"stack"},D={class:"columns"},W={key:0},G={key:1},O=$({__name:"DelegatedGatewayDetailView",setup(Z){return(j,U)=>{const h=r("KCard"),k=r("DataLoader"),f=r("RouterLink"),b=r("KTooltip"),S=r("RouterView"),V=r("AppView"),P=r("RouteView"),T=r("DataSource");return l(),p(T,{src:"/me"},{default:a(({data:z})=>[z?(l(),p(P,{key:0,name:"delegated-gateway-detail-view",params:{mesh:"",service:"",page:1,size:z.pageSize,s:"",dataPlane:""}},{default:a(({can:w,route:t,t:o})=>[n(V,null,{default:a(()=>[u("div",F,[n(k,{src:`/meshes/${t.params.mesh}/service-insights/${t.params.service}`},{default:a(({data:i})=>[i?(l(),p(h,{key:0},{default:a(()=>{var e,d;return[u("div",D,[n(C,null,{title:a(()=>[s(m(o("http.api.property.status")),1)]),body:a(()=>[n(x,{status:i.status},null,8,["status"])]),_:2},1024),s(),n(C,null,{title:a(()=>[s(m(o("http.api.property.address")),1)]),body:a(()=>[i.addressPort?(l(),p(L,{key:0,text:i.addressPort},null,8,["text"])):(l(),c(y,{key:1},[s(m(o("common.detail.none")),1)],64))]),_:2},1024),s(),n(R,{online:((e=i.dataplanes)==null?void 0:e.online)??0,total:((d=i.dataplanes)==null?void 0:d.total)??0},{title:a(()=>[s(m(o("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1024)):_("",!0)]),_:2},1032,["src"]),s(),u("div",null,[u("h3",null,m(o("delegated-gateways.detail.data_plane_proxies")),1),s(),n(h,{class:"mt-4"},{default:a(()=>[n(k,{src:`/meshes/${t.params.mesh}/dataplanes/for/${t.params.service}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`,loader:!1},{default:a(({data:i})=>[n(K,{class:"data-plane-collection","data-testid":"data-plane-collection","page-number":t.params.page,"page-size":t.params.size,headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...w("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:i==null?void 0:i.items,total:i==null?void 0:i.total,"is-selected-row":e=>e.name===t.params.dataPlane,"summary-route-name":"delegated-gateway-data-plane-summary-view","empty-state-message":o("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":o("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":o("common.documentation"),onChange:t.update},{toolbar:a(()=>[n(q,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...w("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>t.update({...Object.fromEntries(e.entries())})},null,8,["placeholder","query","fields","onChange"])]),name:a(({row:e})=>[n(f,{class:"name-link",to:{name:"delegated-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s}}},{default:a(()=>[s(m(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[s(m(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(l(),p(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[s(m(e.zone),1)]),_:2},1032,["to"])):(l(),c(y,{key:1},[s(m(o("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var d;return[(d=e.dataplaneInsight.mTLS)!=null&&d.certificateExpirationTime?(l(),c(y,{key:0},[s(m(o("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(l(),c(y,{key:1},[s(m(o("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[n(x,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(l(),p(b,{key:0},{content:a(()=>[u("ul",null,[e.warnings.length>0?(l(),c("li",W,m(o("data-planes.components.data-plane-list.version_mismatch")),1)):_("",!0),s(),e.isCertExpired?(l(),c("li",G,m(o("data-planes.components.data-plane-list.cert_expired")),1)):_("",!0)])]),default:a(()=>[s(),n(B,{class:"mr-1",size:g(v)},null,8,["size"])]),_:2},1024)):(l(),c(y,{key:1},[s(m(o("common.collection.none")),1)],64))]),details:a(({row:e})=>[n(f,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[s(m(o("common.collection.details_link"))+" ",1),n(g(E),{decorative:"",size:g(v)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]),s(),t.params.dataPlane?(l(),p(S,{key:0},{default:a(e=>[n(A,{onClose:d=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof i<"u"?(l(),p(I(e.Component),{key:0,items:i.items},null,8,["items"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:2},1032,["params"])):_("",!0)]),_:1})}}}),Y=N(O,[["__scopeId","data-v-133cd589"]]);export{Y as default};
