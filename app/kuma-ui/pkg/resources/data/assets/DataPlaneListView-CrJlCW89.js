import{d as E,r as p,o as n,m,w as a,b as o,e as s,k,l as z,ay as X,A as N,R as B,az as P,p as y,t as l,G as b,c as _,F as u,T as $,S as q,E as K,q as F}from"./index-Bs4ZAy4O.js";import{F as G}from"./FilterBar-DT5drdMJ.js";import{S as M}from"./SummaryView-CKUqxdH5.js";const H=["innerHTML"],O={key:0},Z={key:1},j=E({__name:"DataPlaneListView",setup(U){return(W,J)=>{const v=p("RouteTitle"),g=p("XIcon"),w=p("XSelect"),f=p("XAction"),T=p("KTruncate"),C=p("XActionGroup"),x=p("RouterView"),S=p("DataCollection"),V=p("DataLoader"),L=p("KCard"),A=p("AppView"),D=p("RouteView");return n(),m(D,{name:"data-plane-list-view",params:{page:1,size:50,dataplaneType:"all",s:"",mesh:"",dataPlane:""}},{default:a(({can:h,route:t,t:i,me:r,uri:I})=>[o(v,{render:!1,title:i("data-planes.routes.items.title")},null,8,["title"]),s(),o(A,{docs:i("data-planes.href.docs.data_plane_proxy")},{default:a(()=>[k("div",{innerHTML:i("data-planes.routes.items.intro",{},{defaultMessage:""})},null,8,H),s(),o(L,null,{default:a(()=>[o(V,{src:I(z(X),"/meshes/:mesh/dataplanes/of/:type",{mesh:t.params.mesh,type:t.params.dataplaneType},{page:t.params.page,size:t.params.size,search:t.params.s})},{loadable:a(({data:c})=>[o(S,{type:"data-planes",items:(c==null?void 0:c.items)??[void 0]},{default:a(()=>[o(N,{class:"data-plane-collection","data-testid":"data-plane-collection","page-number":t.params.page,"page-size":t.params.size,headers:[{...r.get("headers.type"),label:" ",key:"type"},{...r.get("headers.name"),label:"Name",key:"name"},{...r.get("headers.namespace"),label:"Namespace",key:"namespace"},...h("use zones")?[{...r.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...r.get("headers.services"),label:"Services",key:"services"},{...r.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...r.get("headers.status"),label:"Status",key:"status"},{...r.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...r.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:c==null?void 0:c.items,total:c==null?void 0:c.total,"is-selected-row":e=>e.name===t.params.dataPlane,onChange:t.update,onResize:r.set},{toolbar:a(()=>[o(G,{class:"data-plane-proxy-filter",placeholder:"service:backend",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...h("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>t.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"]),s(),o(w,{label:"Type",selected:t.params.dataplaneType,onChange:e=>t.update({dataplaneType:e})},B({selected:a(({item:e})=>[e!=="all"?(n(),m(g,{key:0,size:z(P),name:e},null,8,["size","name"])):y("",!0),s(" "+l(i(`data-planes.type.${e}`)),1)]),_:2},[b(["all","standard","builtin","delegated"],e=>({name:`${e}-option`,fn:a(()=>[e!=="all"?(n(),m(g,{key:0,name:e},null,8,["name"])):y("",!0),s(" "+l(i(`data-planes.type.${e}`)),1)])}))]),1032,["selected","onChange"])]),type:a(({row:e})=>[o(g,{name:e.dataplaneType},{default:a(()=>[s(l(i(`data-planes.type.${e.dataplaneType}`)),1)]),_:2},1032,["name"])]),name:a(({row:e})=>[o(f,{"data-action":"",class:"name-link",title:e.name,to:{name:"data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s,dataplaneType:t.params.dataplaneType}}},{default:a(()=>[s(l(e.name),1)]),_:2},1032,["title","to"])]),namespace:a(({row:e})=>[s(l(e.namespace),1)]),services:a(({row:e})=>[e.services.length>0?(n(),m(T,{key:0,width:"auto"},{default:a(()=>[(n(!0),_(u,null,b(e.services,(d,R)=>(n(),_("div",{key:R},[o($,{text:d},{default:a(()=>[e.dataplaneType==="standard"?(n(),m(f,{key:0,to:{name:"service-detail-view",params:{service:d}}},{default:a(()=>[s(l(d),1)]),_:2},1032,["to"])):e.dataplaneType==="delegated"?(n(),m(f,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:d}}},{default:a(()=>[s(l(d),1)]),_:2},1032,["to"])):(n(),_(u,{key:2},[s(l(d),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),zone:a(({row:e})=>[e.zone?(n(),m(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[s(l(e.zone),1)]),_:2},1032,["to"])):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var d;return[(d=e.dataplaneInsight.mTLS)!=null&&d.certificateExpirationTime?(n(),_(u,{key:0},[s(l(i("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(n(),_(u,{key:1},[s(l(i("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(q,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(n(),m(g,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[k("ul",null,[e.warnings.length>0?(n(),_("li",O,l(i("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),s(),e.isCertExpired?(n(),_("li",Z,l(i("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(n(),_(u,{key:1},[s(l(i("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(C,null,{default:a(()=>[o(f,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[s(l(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","onChange","onResize"]),s(),o(x,null,{default:a(({Component:e})=>[t.child()?(n(),m(M,{key:0,onClose:d=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof c<"u"?(n(),m(K(e),{key:0,items:c.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])):y("",!0)]),_:2},1024)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}}),ae=F(j,[["__scopeId","data-v-7a75328f"]]);export{ae as default};
