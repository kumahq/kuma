import{d as B,a as r,o as s,b as c,w as t,e as n,m as z,f as o,t as m,c as p,F as u,q as d,p as b,N as E,K as L,E as N,_ as $}from"./index-DPw5bDvs.js";import{A as P}from"./AppCollection-BhcP3aAa.js";import{F as R}from"./FilterBar-DPFEVHTo.js";import{S as q}from"./StatusBadge-CgwF8Xry.js";import{S as T}from"./SummaryView-CT_iUbeB.js";const A={class:"stack"},F={key:0},K={key:1},O=B({__name:"BuiltinGatewayDataplanesView",setup(X){return(Z,j)=>{const _=r("RouterLink"),v=r("XIcon"),w=r("RouterView"),C=r("DataLoader"),x=r("KCard"),g=r("DataSource"),S=r("AppView"),V=r("RouteView");return s(),c(g,{src:"/me"},{default:t(({data:k})=>[k?(s(),c(V,{key:0,name:"builtin-gateway-dataplanes-view",params:{mesh:"",gateway:"",listener:"",page:1,size:k.pageSize,s:"",dataPlane:""}},{default:t(({can:h,route:a,t:i})=>[n(S,null,{default:t(()=>[n(g,{src:`/meshes/${a.params.mesh}/mesh-gateways/${a.params.gateway}`},{default:t(({data:y,error:I})=>[z("div",A,[n(x,null,{default:t(()=>[n(C,{src:y===void 0?"":`/meshes/${a.params.mesh}/dataplanes/for/service-insight/${y.selectors[0].match["kuma.io/service"]}?page=${a.params.page}&size=${a.params.size}&search=${a.params.s}`,data:[y],errors:[I],loader:!1},{default:t(({data:l})=>[n(P,{class:"data-plane-collection","data-testid":"data-plane-collection","page-number":a.params.page,"page-size":a.params.size,headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...h("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:l==null?void 0:l.items,total:l==null?void 0:l.total,"is-selected-row":e=>e.name===a.params.dataPlane,"summary-route-name":"builtin-gateway-data-plane-summary-view","empty-state-message":i("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":i("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":i("common.documentation"),onChange:a.update},{toolbar:t(()=>[n(R,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:a.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...h("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>a.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"])]),namespace:t(({row:e})=>[o(m(e.namespace),1)]),name:t(({row:e})=>[n(_,{class:"name-link",title:e.name,to:{name:"builtin-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:a.params.page,size:a.params.size,s:a.params.s}}},{default:t(()=>[o(m(e.name),1)]),_:2},1032,["title","to"])]),zone:t(({row:e})=>[e.zone?(s(),c(_,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:t(()=>[o(m(e.zone),1)]),_:2},1032,["to"])):(s(),p(u,{key:1},[o(m(i("common.collection.none")),1)],64))]),certificate:t(({row:e})=>{var f;return[(f=e.dataplaneInsight.mTLS)!=null&&f.certificateExpirationTime?(s(),p(u,{key:0},[o(m(i("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(s(),p(u,{key:1},[o(m(i("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:t(({row:e})=>[n(q,{status:e.status},null,8,["status"])]),warnings:t(({row:e})=>[e.isCertExpired||e.warnings.length>0?(s(),c(v,{key:0,class:"mr-1",name:"warning"},{default:t(()=>[z("ul",null,[e.warnings.length>0?(s(),p("li",F,m(i("data-planes.components.data-plane-list.version_mismatch")),1)):d("",!0),o(),e.isCertExpired?(s(),p("li",K,m(i("data-planes.components.data-plane-list.cert_expired")),1)):d("",!0)])]),_:2},1024)):(s(),p(u,{key:1},[o(m(i("common.collection.none")),1)],64))]),details:t(({row:e})=>[n(_,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:t(()=>[o(m(i("common.collection.details_link"))+" ",1),n(b(E),{decorative:"",size:b(L)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"]),o(),a.params.dataPlane?(s(),c(w,{key:0},{default:t(e=>[n(T,{onClose:f=>a.replace({name:a.name,params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size,s:a.params.s}})},{default:t(()=>[typeof l<"u"?(s(),c(N(e.Component),{key:0,items:l.items},null,8,["items"])):d("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):d("",!0)]),_:2},1032,["src","data","errors"])]),_:2},1024)])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["params"])):d("",!0)]),_:1})}}}),J=$(O,[["__scopeId","data-v-79c7b1d8"]]);export{J as default};
