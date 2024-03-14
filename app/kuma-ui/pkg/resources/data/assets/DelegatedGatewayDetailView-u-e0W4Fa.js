import{d as L,a as p,o as l,b as c,w as a,Z as w,t as m,f as s,e as o,W as T,F as u,c as d,av as B,m as _,q as y,U as I,K as C,p as g,V as N,D as R,_ as $}from"./index-CP9JG8i6.js";import{A as E}from"./AppCollection-8qMU5pbA.js";import{F}from"./FilterBar-YHHv_dmB.js";import{S as x}from"./StatusBadge-avZJE2n6.js";import{S as K}from"./SummaryView-X-L7WbN2.js";const A={class:"stack"},W={class:"columns"},Z={key:0},G={key:1},O=L({__name:"DelegatedGatewayDetailView",setup(U){return(D,J)=>{const h=p("KCard"),k=p("DataLoader"),f=p("RouterLink"),b=p("KTooltip"),S=p("RouterView"),V=p("AppView"),P=p("RouteView"),q=p("DataSource");return l(),c(q,{src:"/me"},{default:a(({data:z})=>[z?(l(),c(P,{key:0,name:"delegated-gateway-detail-view",params:{mesh:"",service:"",page:1,size:z.pageSize,query:"",s:"",dataPlane:""}},{default:a(({can:v,route:t,t:i})=>[o(V,null,{default:a(()=>[_("div",A,[o(k,{src:`/meshes/${t.params.mesh}/service-insights/${t.params.service}`},{default:a(({data:n})=>[n?(l(),c(h,{key:0},{default:a(()=>{var e,r;return[_("div",W,[o(w,null,{title:a(()=>[s(m(i("http.api.property.status")),1)]),body:a(()=>[o(x,{status:n.status},null,8,["status"])]),_:2},1024),s(),o(w,null,{title:a(()=>[s(m(i("http.api.property.address")),1)]),body:a(()=>[n.addressPort?(l(),c(T,{key:0,text:n.addressPort},null,8,["text"])):(l(),d(u,{key:1},[s(m(i("common.detail.none")),1)],64))]),_:2},1024),s(),o(B,{online:((e=n.dataplanes)==null?void 0:e.online)??0,total:((r=n.dataplanes)==null?void 0:r.total)??0},{title:a(()=>[s(m(i("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1024)):y("",!0)]),_:2},1032,["src"]),s(),_("div",null,[_("h3",null,m(i("delegated-gateways.detail.data_plane_proxies")),1),s(),o(h,{class:"mt-4"},{default:a(()=>[o(k,{src:`/meshes/${t.params.mesh}/dataplanes/for/${t.params.service}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`,loader:!1},{default:a(({data:n})=>[o(E,{class:"data-plane-collection","data-testid":"data-plane-collection","page-number":t.params.page,"page-size":t.params.size,headers:[{label:"Name",key:"name"},...v("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:n==null?void 0:n.items,total:n==null?void 0:n.total,"is-selected-row":e=>e.name===t.params.dataPlane,"summary-route-name":"delegated-gateway-data-plane-summary-view","empty-state-message":i("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":i("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":i("common.documentation"),onChange:t.update},{toolbar:a(()=>[o(F,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.params.query,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...v("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onFieldsChange:e=>t.update({query:e.query,s:e.query.length>0?JSON.stringify(e.fields):""})},null,8,["placeholder","query","fields","onFieldsChange"])]),name:a(({row:e})=>[o(f,{class:"name-link",title:e.name,to:{name:"delegated-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.name},query:{page:t.params.page,size:t.params.size,query:t.params.query}}},{default:a(()=>[s(m(e.name),1)]),_:2},1032,["title","to"])]),zone:a(({row:e})=>[e.zone?(l(),c(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[s(m(e.zone),1)]),_:2},1032,["to"])):(l(),d(u,{key:1},[s(m(i("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var r;return[(r=e.dataplaneInsight.mTLS)!=null&&r.certificateExpirationTime?(l(),d(u,{key:0},[s(m(i("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(l(),d(u,{key:1},[s(m(i("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(x,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(l(),c(b,{key:0},{content:a(()=>[_("ul",null,[e.warnings.length>0?(l(),d("li",Z,m(i("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),s(),e.isCertExpired?(l(),d("li",G,m(i("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),default:a(()=>[s(),o(I,{class:"mr-1",size:g(C)},null,8,["size"])]),_:2},1024)):(l(),d(u,{key:1},[s(m(i("common.collection.none")),1)],64))]),details:a(({row:e})=>[o(f,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.name}}},{default:a(()=>[s(m(i("common.collection.details_link"))+" ",1),o(g(N),{decorative:"",size:g(C)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]),s(),t.params.dataPlane?(l(),c(S,{key:0},{default:a(e=>[o(K,{onClose:r=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:a(()=>[(l(),c(R(e.Component),{name:t.params.dataPlane,"dataplane-overview":n==null?void 0:n.items.find(r=>r.name===t.params.dataPlane)},null,8,["name","dataplane-overview"]))]),_:2},1032,["onClose"])]),_:2},1024)):y("",!0)]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:2},1032,["params"])):y("",!0)]),_:1})}}}),Y=$(O,[["__scopeId","data-v-efc0a2ef"]]);export{Y as default};
