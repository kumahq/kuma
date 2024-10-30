import{d as P,r as c,o as l,m as p,w as a,b as o,k as u,Z as z,e as n,t as i,S as v,a2 as R,c as d,L as g,$ as S,p as h,A as E,E as B,q as I}from"./index-CMjLgvOo.js";import{F as N}from"./FilterBar-B7m2qAXa.js";import{S as X}from"./SummaryView-C546ionl.js";const q={class:"stack"},T={class:"columns"},G={key:0},F={key:1},K=P({__name:"DelegatedGatewayDetailView",setup(Z){return(j,O)=>{const y=c("KCard"),f=c("DataLoader"),k=c("XAction"),C=c("RouterLink"),b=c("XIcon"),x=c("XActionGroup"),V=c("RouterView"),$=c("DataCollection"),A=c("AppView"),L=c("RouteView");return l(),p(L,{name:"delegated-gateway-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:""}},{default:a(({can:w,route:t,t:r,me:m})=>[o(A,null,{default:a(()=>[u("div",q,[o(f,{src:`/meshes/${t.params.mesh}/service-insights/${t.params.service}`},{default:a(({data:s})=>[s?(l(),p(y,{key:0},{default:a(()=>{var e,_;return[u("div",T,[o(z,null,{title:a(()=>[n(i(r("http.api.property.status")),1)]),body:a(()=>[o(v,{status:s.status},null,8,["status"])]),_:2},1024),n(),o(z,null,{title:a(()=>[n(i(r("http.api.property.address")),1)]),body:a(()=>[s.addressPort?(l(),p(R,{key:0,text:s.addressPort},null,8,["text"])):(l(),d(g,{key:1},[n(i(r("common.detail.none")),1)],64))]),_:2},1024),n(),o(S,{online:((e=s.dataplanes)==null?void 0:e.online)??0,total:((_=s.dataplanes)==null?void 0:_.total)??0},{title:a(()=>[n(i(r("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1024)):h("",!0)]),_:2},1032,["src"]),n(),u("div",null,[u("h3",null,i(r("delegated-gateways.detail.data_plane_proxies")),1),n(),o(y,{class:"mt-4"},{default:a(()=>[u("search",null,[o(N,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...w("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:s=>t.update({...Object.fromEntries(s.entries())})},null,8,["query","fields","onChange"])]),n(),o(f,{src:`/meshes/${t.params.mesh}/dataplanes/for/service-insight/${t.params.service}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`},{loadable:a(({data:s})=>[o($,{type:"data-planes",items:(s==null?void 0:s.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:s==null?void 0:s.total,onChange:t.update},{default:a(()=>[o(E,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.namespace"),label:"Namespace",key:"namespace"},...w("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...m.get("headers.status"),label:"Status",key:"status"},{...m.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,"is-selected-row":e=>e.name===t.params.dataPlane,onResize:m.set},{name:a(({row:e})=>[o(k,{"data-action":"",class:"name-link",to:{name:"delegated-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s}}},{default:a(()=>[n(i(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(i(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(l(),p(C,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(i(e.zone),1)]),_:2},1032,["to"])):(l(),d(g,{key:1},[n(i(r("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var _;return[(_=e.dataplaneInsight.mTLS)!=null&&_.certificateExpirationTime?(l(),d(g,{key:0},[n(i(r("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(l(),d(g,{key:1},[n(i(r("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(v,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(l(),p(b,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[u("ul",null,[e.warnings.length>0?(l(),d("li",G,i(r("data-planes.components.data-plane-list.version_mismatch")),1)):h("",!0),n(),e.isCertExpired?(l(),d("li",F,i(r("data-planes.components.data-plane-list.cert_expired")),1)):h("",!0)])]),_:2},1024)):(l(),d(g,{key:1},[n(i(r("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(x,null,{default:a(()=>[o(k,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[n(i(r("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),n(),t.params.dataPlane?(l(),p(V,{key:0},{default:a(e=>[o(X,{onClose:_=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof s<"u"?(l(),p(B(e.Component),{key:0,items:s.items},null,8,["items"])):h("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):h("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),M=I(K,[["__scopeId","data-v-ae07f386"]]);export{M as default};
