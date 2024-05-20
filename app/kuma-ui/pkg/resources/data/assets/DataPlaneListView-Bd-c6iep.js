import{d as D,a as r,o as n,b as i,w as a,e as p,m as v,f as s,$ as L,Q as P,p as k,ae as B,q as y,t as l,G as C,c,F as u,T as N,O as R,K,E as q,_ as A}from"./index-Dlb3STpa.js";import{A as F}from"./AppCollection-DFkcV_Z2.js";import{F as O}from"./FilterBar-CUrz5nnb.js";import{S as X}from"./StatusBadge-Dt3Uz7hD.js";import{S as Z}from"./SummaryView-BtC60_FZ.js";const U={key:0},W={key:1},j=D({__name:"DataPlaneListView",setup(G){return(Q,H)=>{const T=r("RouteTitle"),f=r("XIcon"),w=r("XSelect"),_=r("RouterLink"),x=r("KTruncate"),S=r("RouterView"),V=r("KCard"),I=r("AppView"),h=r("DataSource"),$=r("RouteView");return n(),i(h,{src:"/me"},{default:a(({data:z})=>[z?(n(),i($,{key:0,name:"data-plane-list-view",params:{page:1,size:z.pageSize,dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:a(({can:b,route:t,t:o})=>[p(h,{src:`/meshes/${t.params.mesh}/dataplanes/of/${t.params.dataplaneType}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`},{default:a(({data:d,error:g})=>[p(I,null,{title:a(()=>[v("h2",null,[p(T,{title:o("data-planes.routes.items.title")},null,8,["title"])])]),default:a(()=>[s(),p(V,null,{default:a(()=>[g!==void 0?(n(),i(L,{key:0,error:g},null,8,["error"])):(n(),i(F,{key:1,class:"data-plane-collection","data-testid":"data-plane-collection","page-number":t.params.page,"page-size":t.params.size,headers:[{label:" ",key:"type"},{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...b("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Services",key:"services"},{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],items:d==null?void 0:d.items,total:d==null?void 0:d.total,error:g,"is-selected-row":e=>e.name===t.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":o("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":o("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":o("common.documentation"),onChange:t.update},{toolbar:a(()=>[p(O,{class:"data-plane-proxy-filter",placeholder:"service:backend",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...b("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>t.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"]),s(),p(w,{label:"Type",selected:t.params.dataplaneType,onChange:e=>t.update({dataplaneType:e})},P({selected:a(({item:e})=>[e!=="all"?(n(),i(f,{key:0,size:k(B),name:e},null,8,["size","name"])):y("",!0),s(" "+l(o(`data-planes.type.${e}`)),1)]),_:2},[C(["all","standard","builtin","delegated"],e=>({name:`${e}-option`,fn:a(()=>[e!=="all"?(n(),i(f,{key:0,name:e},null,8,["name"])):y("",!0),s(" "+l(o(`data-planes.type.${e}`)),1)])}))]),1032,["selected","onChange"])]),type:a(({row:e})=>[p(f,{name:e.dataplaneType},{default:a(()=>[s(l(o(`data-planes.type.${e.dataplaneType}`)),1)]),_:2},1032,["name"])]),name:a(({row:e})=>[p(_,{"data-action":"",class:"name-link",title:e.name,to:{name:"data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s,dataplaneType:t.params.dataplaneType}}},{default:a(()=>[s(l(e.name),1)]),_:2},1032,["title","to"])]),namespace:a(({row:e})=>[s(l(e.namespace),1)]),services:a(({row:e})=>[e.services.length>0?(n(),i(x,{key:0,width:"auto"},{default:a(()=>[(n(!0),c(u,null,C(e.services,(m,E)=>(n(),c("div",{key:E},[p(N,{text:m},{default:a(()=>[e.dataplaneType==="standard"?(n(),i(_,{key:0,to:{name:"service-detail-view",params:{service:m}}},{default:a(()=>[s(l(m),1)]),_:2},1032,["to"])):e.dataplaneType==="delegated"?(n(),i(_,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:m}}},{default:a(()=>[s(l(m),1)]),_:2},1032,["to"])):(n(),c(u,{key:2},[s(l(m),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(n(),c(u,{key:1},[s(l(o("common.collection.none")),1)],64))]),zone:a(({row:e})=>[e.zone?(n(),i(_,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[s(l(e.zone),1)]),_:2},1032,["to"])):(n(),c(u,{key:1},[s(l(o("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var m;return[(m=e.dataplaneInsight.mTLS)!=null&&m.certificateExpirationTime?(n(),c(u,{key:0},[s(l(o("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(n(),c(u,{key:1},[s(l(o("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[p(X,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(n(),i(f,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[v("ul",null,[e.warnings.length>0?(n(),c("li",U,l(o("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),s(),e.isCertExpired?(n(),c("li",W,l(o("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(n(),c(u,{key:1},[s(l(o("common.collection.none")),1)],64))]),details:a(({row:e})=>[p(_,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[s(l(o("common.collection.details_link"))+" ",1),p(k(R),{decorative:"",size:k(K)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["page-number","page-size","headers","items","total","error","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange"])),s(),t.params.dataPlane?(n(),i(S,{key:2},{default:a(e=>[p(Z,{onClose:m=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof d<"u"?(n(),i(q(e.Component),{key:0,items:d.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):y("",!0)]),_:2},1024)]),_:2},1024)]),_:2},1032,["src"])]),_:2},1032,["params"])):y("",!0)]),_:1})}}}),te=A(j,[["__scopeId","data-v-d75f8ebf"]]);export{te as default};
