import{d as K,r as i,m as y,o as s,w as a,b as l,e as o,q as x,M as P,p as b,s as X,a3 as Z,t as p,v,O as j,C as M,c,F as f,S as U,K as W,_ as H}from"./index-BfbnxhY1.js";import{F as J}from"./FilterBar-BR-Uthpf.js";import{S as Q}from"./SummaryView-CK2XSYPn.js";const Y=["data-testid"],ee=K({__name:"DataPlaneListView",props:{mesh:{}},setup(I){const S=I;return(ae,d)=>{const V=i("RouteTitle"),D=i("XI18n"),h=i("XIcon"),L=i("XSelect"),_=i("XAction"),A=i("XCopyButton"),B=i("XLayout"),N=i("XActionGroup"),R=i("RouterView"),$=i("DataCollection"),E=i("DataLoader"),q=i("XCard"),F=i("AppView"),O=i("RouteView");return s(),y(O,{name:"data-plane-list-view",params:{page:1,size:Number,dataplaneType:"all",s:"",mesh:"",proxy:""}},{default:a(({can:C,route:t,t:m,me:u,uri:G})=>[l(V,{render:!1,title:m("data-planes.routes.items.title")},null,8,["title"]),d[13]||(d[13]=o()),l(F,{docs:m("data-planes.href.docs.data_plane_proxy")},{default:a(()=>[l(D,{path:"data-planes.routes.items.intro","default-path":"common.i18n.ignore-error"}),d[12]||(d[12]=o()),l(q,null,{default:a(()=>[x("search",null,[l(J,{class:"data-plane-proxy-filter",placeholder:"service:backend",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...C("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:n=>t.update({page:1,...Object.fromEntries(n.entries())})},null,8,["query","fields","onChange"]),d[1]||(d[1]=o()),l(L,{label:"Type",selected:t.params.dataplaneType,onChange:n=>t.update({page:1,dataplaneType:n})},P({selected:a(({item:n})=>[n!=="all"?(s(),y(h,{key:0,size:X(Z),name:n},null,8,["size","name"])):b("",!0),o(" "+p(m(`data-planes.type.${n}`)),1)]),_:2},[v(["all","standard","builtin","delegated"],n=>({name:`${n}-option`,fn:a(()=>[n!=="all"?(s(),y(h,{key:0,name:n},null,8,["name"])):b("",!0),o(" "+p(m(`data-planes.type.${n}`)),1)])}))]),1032,["selected","onChange"])]),d[11]||(d[11]=o()),l(E,{src:G(X(j),"/meshes/:mesh/dataplanes/of/:type",{mesh:t.params.mesh,type:t.params.dataplaneType},{page:t.params.page,size:t.params.size,search:t.params.s})},{loadable:a(({data:n})=>[l($,{type:"data-planes",items:(n==null?void 0:n.items)??[void 0],total:n==null?void 0:n.total,page:t.params.page,"page-size":t.params.size,onChange:t.update},{default:a(()=>[l(M,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...u.get("headers.type"),label:" ",key:"type"},{...u.get("headers.name"),label:"Name",key:"name"},{...u.get("headers.namespace"),label:"Namespace",key:"namespace"},...C("use zones")?[{...u.get("headers.zone"),label:"Zone",key:"zone"}]:[],...C("use service-insights",S.mesh)?[{...u.get("headers.services"),label:"Services",key:"services"}]:[],{...u.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...u.get("headers.status"),label:"Status",key:"status"},{...u.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...u.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:n==null?void 0:n.items,"is-selected-row":e=>e.name===t.params.proxy,onResize:u.set},{type:a(({row:e})=>[l(h,{name:e.dataplaneType},{default:a(()=>[o(p(m(`data-planes.type.${e.dataplaneType}`)),1)]),_:2},1032,["name"])]),name:a(({row:e})=>[l(_,{"data-action":"",class:"name-link",title:e.name,to:{name:"data-plane-summary-view",params:{mesh:e.mesh,proxy:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s,dataplaneType:t.params.dataplaneType}}},{default:a(()=>[o(p(e.name),1)]),_:2},1032,["title","to"])]),namespace:a(({row:e})=>[o(p(e.namespace),1)]),services:a(({row:e})=>[e.services.length>0?(s(),y(B,{key:0,type:"separated",truncate:""},{default:a(()=>[(s(!0),c(f,null,v(e.services,(r,k)=>(s(),c("div",{key:k},[l(A,{text:r},{default:a(()=>[e.dataplaneType==="standard"?(s(),y(_,{key:0,to:{name:"service-detail-view",params:{service:r}}},{default:a(()=>[o(p(r),1)]),_:2},1032,["to"])):e.dataplaneType==="delegated"?(s(),y(_,{key:1,to:{name:"delegated-gateway-detail-view",params:{service:r}}},{default:a(()=>[o(p(r),1)]),_:2},1032,["to"])):(s(),c(f,{key:2},[o(p(r),1)],64))]),_:2},1032,["text"])]))),128))]),_:2},1024)):(s(),c(f,{key:1},[o(p(m("common.collection.none")),1)],64))]),zone:a(({row:e})=>[e.zone?(s(),y(_,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[o(p(e.zone),1)]),_:2},1032,["to"])):(s(),c(f,{key:1},[o(p(m("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var r;return[(r=e.dataplaneInsight.mTLS)!=null&&r.certificateExpirationTime?(s(),c(f,{key:0},[o(p(m("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(s(),c(f,{key:1},[o(p(m("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[l(U,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>{var r,k,w,T;return[(s(!0),c(f,null,v([[{bool:((k=(r=e.dataplaneInsight.version)==null?void 0:r.kumaDp)==null?void 0:k.kumaCpCompatible)===!1||((T=(w=e.dataplaneInsight.version)==null?void 0:w.envoy)==null?void 0:T.kumaDpCompatible)===!1,key:"dp-cp-incompatible"},{bool:e.isCertExpired,key:"certificate-expired"}].filter(({bool:g})=>g)],g=>(s(),c(f,{key:typeof g},[g.length>0?(s(),y(h,{key:0,name:"warning","data-testid":"warning"},{default:a(()=>[x("ul",null,[(s(!0),c(f,null,v(g,({key:z})=>(s(),c("li",{key:z,"data-testid":`warning-${z}`},p(m(`data-planes.routes.items.warnings.${z}`)),9,Y))),128))])]),_:2},1024)):(s(),c(f,{key:1},[o(p(m("common.collection.none")),1)],64))],64))),128))]}),actions:a(({row:e})=>[l(N,null,{default:a(()=>[l(_,{to:{name:"data-plane-detail-view",params:{proxy:e.id}}},{default:a(()=>[o(p(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),d[10]||(d[10]=o()),l(R,null,{default:a(({Component:e})=>[t.child()?(s(),y(Q,{key:0,onClose:r=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof n<"u"?(s(),y(W(e),{key:0,items:n.items},null,8,["items"])):b("",!0)]),_:2},1032,["onClose"])):b("",!0)]),_:2},1024)]),_:2},1032,["items","total","page","page-size","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}}),oe=H(ee,[["__scopeId","data-v-e4235ba5"]]);export{oe as default};
