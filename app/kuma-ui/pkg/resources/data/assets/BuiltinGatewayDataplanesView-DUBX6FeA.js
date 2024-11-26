import{d as E,e as l,o as i,m as u,w as s,a as n,k,b as o,A as I,t as r,c as d,J as _,S as L,p as g,F as N,q as R}from"./index-CKcsX_-l.js";import{F as X}from"./FilterBar-DdKAp1jk.js";import{S as q}from"./SummaryView-BIwsKbzL.js";const P={class:"stack"},T={key:0},F={key:1},G=E({__name:"BuiltinGatewayDataplanesView",setup(K){return(j,p)=>{const f=l("XAction"),v=l("XIcon"),w=l("XActionGroup"),C=l("RouterView"),b=l("DataCollection"),V=l("DataLoader"),x=l("KCard"),S=l("DataSource"),A=l("AppView"),B=l("RouteView");return i(),u(B,{name:"builtin-gateway-dataplanes-view",params:{mesh:"",gateway:"",listener:"",page:1,size:50,s:"",dataPlane:""}},{default:s(({can:z,route:a,t:c,me:m})=>[n(A,null,{default:s(()=>[n(S,{src:`/meshes/${a.params.mesh}/mesh-gateways/${a.params.gateway}`},{default:s(({data:y,error:$})=>[k("div",P,[n(x,null,{default:s(()=>[k("search",null,[n(X,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:a.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...z("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:t=>a.update({...Object.fromEntries(t.entries())})},null,8,["query","fields","onChange"])]),p[8]||(p[8]=o()),n(V,{src:y===void 0?"":`/meshes/${a.params.mesh}/dataplanes/for/service-insight/${y.selectors[0].match["kuma.io/service"]}?page=${a.params.page}&size=${a.params.size}&search=${a.params.s}`,data:[y],errors:[$]},{loadable:s(({data:t})=>[n(b,{type:"data-planes",items:(t==null?void 0:t.items)??[void 0],total:t==null?void 0:t.total,page:a.params.page,"page-size":a.params.size,onChange:a.update},{default:s(()=>[n(I,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.namespace"),label:"Namespace",key:"namespace"},...z("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...m.get("headers.status"),label:"Status",key:"status"},{...m.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":e=>e.name===a.params.dataPlane,onResize:m.set},{namespace:s(({row:e})=>[o(r(e.namespace),1)]),name:s(({row:e})=>[n(f,{"data-action":"",class:"name-link",title:e.name,to:{name:"builtin-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:a.params.page,size:a.params.size,s:a.params.s}}},{default:s(()=>[o(r(e.name),1)]),_:2},1032,["title","to"])]),zone:s(({row:e})=>[e.zone?(i(),u(f,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[o(r(e.zone),1)]),_:2},1032,["to"])):(i(),d(_,{key:1},[o(r(c("common.collection.none")),1)],64))]),certificate:s(({row:e})=>{var h;return[(h=e.dataplaneInsight.mTLS)!=null&&h.certificateExpirationTime?(i(),d(_,{key:0},[o(r(c("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(i(),d(_,{key:1},[o(r(c("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:s(({row:e})=>[n(L,{status:e.status},null,8,["status"])]),warnings:s(({row:e})=>[e.isCertExpired||e.warnings.length>0?(i(),u(v,{key:0,class:"mr-1",name:"warning"},{default:s(()=>[k("ul",null,[e.warnings.length>0?(i(),d("li",T,r(c("data-planes.components.data-plane-list.version_mismatch")),1)):g("",!0),p[0]||(p[0]=o()),e.isCertExpired?(i(),d("li",F,r(c("data-planes.components.data-plane-list.cert_expired")),1)):g("",!0)])]),_:2},1024)):(i(),d(_,{key:1},[o(r(c("common.collection.none")),1)],64))]),actions:s(({row:e})=>[n(w,null,{default:s(()=>[n(f,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:s(()=>[o(r(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),p[7]||(p[7]=o()),a.params.dataPlane?(i(),u(C,{key:0},{default:s(e=>[n(q,{onClose:h=>a.replace({name:a.name,params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size,s:a.params.s}})},{default:s(()=>[typeof t<"u"?(i(),u(N(e.Component),{key:0,items:t.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["items","total","page","page-size","onChange"])]),_:2},1032,["src","data","errors"])]),_:2},1024)])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}}),Z=R(G,[["__scopeId","data-v-07564e01"]]);export{Z as default};
