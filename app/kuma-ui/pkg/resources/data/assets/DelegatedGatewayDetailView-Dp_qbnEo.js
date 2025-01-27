import{d as $,r as p,o as r,q as d,w as a,b as o,U as z,e as n,t as i,S as C,c as u,M as y,V as I,s as _,m as f,B as R,I as T,_ as E}from"./index-oTPgN0we.js";import{F as N}from"./FilterBar-B6pvid3Q.js";import{S as q}from"./SummaryView-DipvUpKS.js";const G={key:0},F={key:1},j=$({__name:"DelegatedGatewayDetailView",setup(M){return(O,l)=>{const b=p("XCopyButton"),v=p("XAboutCard"),k=p("DataLoader"),h=p("XAction"),x=p("XIcon"),V=p("XActionGroup"),X=p("RouterView"),A=p("DataCollection"),B=p("XCard"),P=p("XLayout"),S=p("AppView"),L=p("RouteView");return r(),d(L,{name:"delegated-gateway-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:""}},{default:a(({can:w,route:t,t:m,me:c})=>[o(S,null,{default:a(()=>[o(P,{type:"stack"},{default:a(()=>[o(k,{src:`/meshes/${t.params.mesh}/service-insights/${t.params.service}`},{default:a(({data:s})=>[s?(r(),d(v,{key:0,title:m("delegated-gateways.detail.about.title"),created:s.creationTime,modified:s.modificationTime},{default:a(()=>{var e,g;return[o(z,{layout:"horizontal"},{title:a(()=>[n(i(m("http.api.property.status")),1)]),body:a(()=>[o(C,{status:s.status},null,8,["status"])]),_:2},1024),l[2]||(l[2]=n()),o(z,{layout:"horizontal"},{title:a(()=>[n(i(m("http.api.property.address")),1)]),body:a(()=>[s.addressPort?(r(),d(b,{key:0,variant:"badge",format:"default",text:s.addressPort},null,8,["text"])):(r(),u(y,{key:1},[n(i(m("common.detail.none")),1)],64))]),_:2},1024),l[3]||(l[3]=n()),o(I,{layout:"horizontal",online:((e=s.dataplanes)==null?void 0:e.online)??0,total:((g=s.dataplanes)==null?void 0:g.total)??0},{title:a(()=>[n(i(m("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])]}),_:2},1032,["title","created","modified"])):_("",!0)]),_:2},1032,["src"]),l[14]||(l[14]=n()),f("div",null,[f("h3",null,i(m("delegated-gateways.detail.data_plane_proxies")),1),l[13]||(l[13]=n()),o(B,{class:"mt-4"},{default:a(()=>[f("search",null,[o(N,{class:"data-plane-proxy-filter",placeholder:"tag: 'kuma.io/protocol: http'",query:t.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...w("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:s=>t.update({...Object.fromEntries(s.entries())})},null,8,["query","fields","onChange"])]),l[12]||(l[12]=n()),o(k,{src:`/meshes/${t.params.mesh}/dataplanes/for/service-insight/${t.params.service}?page=${t.params.page}&size=${t.params.size}&search=${t.params.s}`},{loadable:a(({data:s})=>[o(A,{type:"data-planes",items:(s==null?void 0:s.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:s==null?void 0:s.total,onChange:t.update},{default:a(()=>[o(R,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.namespace"),label:"Namespace",key:"namespace"},...w("use zones")?[{...c.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...c.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,"is-selected-row":e=>e.name===t.params.dataPlane,onResize:c.set},{name:a(({row:e})=>[o(h,{"data-action":"",class:"name-link",to:{name:"delegated-gateway-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:t.params.page,size:t.params.size,s:t.params.s}}},{default:a(()=>[n(i(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(i(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(r(),d(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(i(e.zone),1)]),_:2},1032,["to"])):(r(),u(y,{key:1},[n(i(m("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var g;return[(g=e.dataplaneInsight.mTLS)!=null&&g.certificateExpirationTime?(r(),u(y,{key:0},[n(i(m("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(r(),u(y,{key:1},[n(i(m("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(C,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(r(),d(x,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[f("ul",null,[e.warnings.length>0?(r(),u("li",G,i(m("data-planes.components.data-plane-list.version_mismatch")),1)):_("",!0),l[4]||(l[4]=n()),e.isCertExpired?(r(),u("li",F,i(m("data-planes.components.data-plane-list.cert_expired")),1)):_("",!0)])]),_:2},1024)):(r(),u(y,{key:1},[n(i(m("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(V,null,{default:a(()=>[o(h,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[n(i(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l[11]||(l[11]=n()),t.params.dataPlane?(r(),d(X,{key:0},{default:a(e=>[o(q,{onClose:g=>t.replace({name:t.name,params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size,s:t.params.s}})},{default:a(()=>[typeof s<"u"?(r(),d(T(e.Component),{key:0,items:s.items},null,8,["items"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1024)]),_:1})}}}),H=E(j,[["__scopeId","data-v-f7a6424c"]]);export{H as default};
