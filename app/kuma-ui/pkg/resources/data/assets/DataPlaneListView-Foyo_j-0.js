import{d as L,a as c,o as s,b as m,w as a,e as o,m as v,f as n,E as $,t as l,c as d,F as y,G as B,T as R,q as _,P as I,p as g,K as T,Q as N,D as q,_ as A}from"./index-zCwNwsfB.js";import{A as F}from"./AppCollection-BDYgDHk6.js";import{F as O}from"./FilterBar-DGcopdSQ.js";import{S as W}from"./StatusBadge-vrODdosP.js";import{S as Z}from"./SummaryView-C1nFaycj.js";const j={key:0},G={key:1},Q=L({__name:"DataPlaneListView",setup(U){return(H,J)=>{const C=c("RouteTitle"),w=c("KSelect"),u=c("RouterLink"),S=c("KTruncate"),x=c("KTooltip"),V=c("RouterView"),D=c("KCard"),K=c("AppView"),k=c("DataSource"),P=c("RouteView");return s(),m(k,{src:"/me"},{default:a(({data:h})=>[h?(s(),m(P,{key:0,name:"data-plane-list-view",params:{page:1,size:h.pageSize,dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:a(({can:b,route:t,t:i})=>[o(k,{src:`/meshes/${t.params.mesh}/dataplanes/of/${t.params.dataplaneType}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`},{default:a(({data:p,error:f})=>[o(K,null,{title:a(()=>[v("h2",null,[o(C,{title:i("data-planes.routes.items.title")},null,8,["title"])])]),default:a(()=>[n(),o(D,null,{default:a(()=>{var z;return[f!==void 0?(s(),m($,{key:0,error:f},null,8,["error"])):(s(),m(F,{key:1,class:"data-plane-collection","data-testid":"data-plane-collection","page-number":t.params.page,"page-size":t.params.size,headers:[{label:"Name",key:"name"},...(((z=p==null?void 0:p.items[0])==null?void 0:z.namespace)??"").length>0?[{label:"Namespace",key:"namespace"}]:[],{label:"Type",key:"type"},{label:"Services",key:"services"},...b("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:p==null?void 0:p.items,total:p==null?void 0:p.total,error:f,"is-selected-row":e=>e.name===t.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":i("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":i("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":i("common.documentation"),onChange:t.update},{toolbar:a(()=>[o(O,{class:"data-plane-proxy-filter",placeholder:"service:backend",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...b("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>t.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"]),n(),o(w,{class:"filter-select",label:"Type",items:["all","standard","builtin","delegated"].map(e=>({value:e,label:i(`data-planes.type.${e}`),selected:e===t.params.dataplaneType})),onSelected:e=>t.update({dataplaneType:String(e.value)})},{"item-template":a(({item:e})=>[n(l(e.label),1)]),_:2},1032,["items","onSelected"])]),name:a(({row:e})=>[o(u,{class:"name-link",title:e.name,to:{name:"data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s,dataplaneType:t.params.dataplaneType}}},{default:a(()=>[n(l(e.name),1)]),_:2},1032,["title","to"])]),namespace:a(({row:e})=>[n(l(e.namespace),1)]),type:a(({row:e})=>[n(l(i(`data-planes.type.${e.dataplaneType}`)),1)]),services:a(({row:e})=>[e.services.length>0?(s(),m(S,{key:0,width:"auto"},{default:a(()=>[(s(!0),d(y,null,B(e.services,(r,E)=>(s(),d("div",{key:E},[o(R,{text:r},{default:a(()=>[e.dataplaneType==="standard"?(s(),m(u,{key:0,to:{name:"service-detail-view",params:{service:r}}},{default:a(()=>[n(l(r),1)]),_:2},1032,["to"])):e.dataplaneType==="delegated"?(s(),m(u,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:r}}},{default:a(()=>[n(l(r),1)]),_:2},1032,["to"])):(s(),d(y,{key:2},[n(l(r),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(s(),d(y,{key:1},[n(l(i("common.collection.none")),1)],64))]),zone:a(({row:e})=>[e.zone?(s(),m(u,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(l(e.zone),1)]),_:2},1032,["to"])):(s(),d(y,{key:1},[n(l(i("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var r;return[(r=e.dataplaneInsight.mTLS)!=null&&r.certificateExpirationTime?(s(),d(y,{key:0},[n(l(i("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(s(),d(y,{key:1},[n(l(i("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(W,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(s(),m(x,{key:0},{content:a(()=>[v("ul",null,[e.warnings.length>0?(s(),d("li",j,l(i("data-planes.components.data-plane-list.version_mismatch")),1)):_("",!0),n(),e.isCertExpired?(s(),d("li",G,l(i("data-planes.components.data-plane-list.cert_expired")),1)):_("",!0)])]),default:a(()=>[n(),o(I,{class:"mr-1",size:g(T)},null,8,["size"])]),_:2},1024)):(s(),d(y,{key:1},[n(l(i("common.collection.none")),1)],64))]),details:a(({row:e})=>[o(u,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.name}}},{default:a(()=>[n(l(i("common.collection.details_link"))+" ",1),o(g(N),{decorative:"",size:g(T)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","error","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])),n(),t.params.dataPlane?(s(),m(V,{key:2},{default:a(e=>[o(Z,{onClose:r=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof p<"u"?(s(),m(q(e.Component),{key:0,items:p.items},null,8,["items"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]}),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):_("",!0)]),_:1})}}}),te=A(Q,[["__scopeId","data-v-bbe5b2e8"]]);export{te as default};
