import{d as N,e as p,o as n,m,w as a,a as o,b as s,k,l as v,a5 as $,A as B,Q as P,ai as q,p as y,t as l,J as z,c as _,H as u,$ as K,S as F,E as H,q as M}from"./index-Yqc5mH7h.js";import{F as G}from"./FilterBar-DhRiSa74.js";import{S as O}from"./SummaryView-DYHwfcRA.js";const Z=["innerHTML"],j={key:0},J={key:1},Q=N({__name:"DataPlaneListView",props:{mesh:{}},setup(b){const w=b;return(U,W)=>{const C=p("RouteTitle"),f=p("XIcon"),T=p("XSelect"),g=p("XAction"),x=p("KTruncate"),S=p("XActionGroup"),V=p("RouterView"),L=p("DataCollection"),A=p("DataLoader"),D=p("KCard"),I=p("AppView"),E=p("RouteView");return n(),m(E,{name:"data-plane-list-view",params:{page:1,size:50,dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:a(({can:h,route:t,t:i,me:r,uri:R})=>[o(C,{render:!1,title:i("data-planes.routes.items.title")},null,8,["title"]),s(),o(I,{docs:i("data-planes.href.docs.data_plane_proxy")},{default:a(()=>[k("div",{innerHTML:i("data-planes.routes.items.intro",{},{defaultMessage:""})},null,8,Z),s(),o(D,null,{default:a(()=>[o(A,{src:R(v($),"/meshes/:mesh/dataplanes/of/:type",{mesh:t.params.mesh,type:t.params.dataplaneType},{page:t.params.page,size:t.params.size,search:t.params.s})},{loadable:a(({data:c})=>[o(L,{type:"data-planes",items:(c==null?void 0:c.items)??[void 0],total:c==null?void 0:c.total,page:t.params.page,"page-size":t.params.size,onChange:t.update},{default:a(()=>[o(B,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...r.get("headers.type"),label:" ",key:"type"},{...r.get("headers.name"),label:"Name",key:"name"},{...r.get("headers.namespace"),label:"Namespace",key:"namespace"},...h("use zones")?[{...r.get("headers.zone"),label:"Zone",key:"zone"}]:[],...h("use service-insights",w.mesh)?[{...r.get("headers.services"),label:"Services",key:"services"}]:[],{...r.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...r.get("headers.status"),label:"Status",key:"status"},{...r.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...r.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:c==null?void 0:c.items,"is-selected-row":e=>e.name===t.params.dataPlane,onResize:r.set},{toolbar:a(()=>[o(G,{class:"data-plane-proxy-filter",placeholder:"service:backend",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...h("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>t.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"]),s(),o(T,{label:"Type",selected:t.params.dataplaneType,onChange:e=>t.update({dataplaneType:e})},P({selected:a(({item:e})=>[e!=="all"?(n(),m(f,{key:0,size:v(q),name:e},null,8,["size","name"])):y("",!0),s(" "+l(i(`data-planes.type.${e}`)),1)]),_:2},[z(["all","standard","builtin","delegated"],e=>({name:`${e}-option`,fn:a(()=>[e!=="all"?(n(),m(f,{key:0,name:e},null,8,["name"])):y("",!0),s(" "+l(i(`data-planes.type.${e}`)),1)])}))]),1032,["selected","onChange"])]),type:a(({row:e})=>[o(f,{name:e.dataplaneType},{default:a(()=>[s(l(i(`data-planes.type.${e.dataplaneType}`)),1)]),_:2},1032,["name"])]),name:a(({row:e})=>[o(g,{"data-action":"",class:"name-link",title:e.name,to:{name:"data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s,dataplaneType:t.params.dataplaneType}}},{default:a(()=>[s(l(e.name),1)]),_:2},1032,["title","to"])]),namespace:a(({row:e})=>[s(l(e.namespace),1)]),services:a(({row:e})=>[e.services.length>0?(n(),m(x,{key:0,width:"auto"},{default:a(()=>[(n(!0),_(u,null,z(e.services,(d,X)=>(n(),_("div",{key:X},[o(K,{text:d},{default:a(()=>[e.dataplaneType==="standard"?(n(),m(g,{key:0,to:{name:"service-detail-view",params:{service:d}}},{default:a(()=>[s(l(d),1)]),_:2},1032,["to"])):e.dataplaneType==="delegated"?(n(),m(g,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:d}}},{default:a(()=>[s(l(d),1)]),_:2},1032,["to"])):(n(),_(u,{key:2},[s(l(d),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),zone:a(({row:e})=>[e.zone?(n(),m(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[s(l(e.zone),1)]),_:2},1032,["to"])):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var d;return[(d=e.dataplaneInsight.mTLS)!=null&&d.certificateExpirationTime?(n(),_(u,{key:0},[s(l(i("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(n(),_(u,{key:1},[s(l(i("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(F,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(n(),m(f,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[k("ul",null,[e.warnings.length>0?(n(),_("li",j,l(i("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),s(),e.isCertExpired?(n(),_("li",J,l(i("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(S,null,{default:a(()=>[o(g,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[s(l(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),s(),o(V,null,{default:a(({Component:e})=>[t.child()?(n(),m(O,{key:0,onClose:d=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof c<"u"?(n(),m(H(e),{key:0,items:c.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])):y("",!0)]),_:2},1024)]),_:2},1032,["items","total","page","page-size","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}}),te=M(Q,[["__scopeId","data-v-db7ebd6e"]]);export{te as default};
