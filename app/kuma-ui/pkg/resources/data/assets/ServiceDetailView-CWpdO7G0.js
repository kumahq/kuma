import{d as I,r as m,o as p,q as f,w as a,b as o,m as _,p as b,aq as N,U as C,e as n,t as i,S as w,c as d,M as y,V as T,Y as q,B as L,s as g,I as F,_ as P}from"./index-BP47cGGe.js";import{F as G}from"./FilterBar-yXS4PZLc.js";import{S as $}from"./SummaryView-Cb9BKz7Y.js";const j={class:"stack"},M={key:0},O={key:1},U=I({__name:"ServiceDetailView",setup(W){return(Y,l)=>{const x=m("XCopyButton"),V=m("XAboutCard"),v=m("DataLoader"),h=m("XAction"),S=m("XIcon"),X=m("XActionGroup"),A=m("RouterView"),D=m("DataCollection"),B=m("XCard"),R=m("AppView"),E=m("RouteView");return p(),f(E,{name:"service-detail-view",params:{mesh:"",service:"",page:1,size:Number,s:"",proxy:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({can:z,route:s,t:r,me:c,uri:k})=>[o(R,null,{default:a(()=>[_("div",j,[o(v,{src:k(b(N),"/meshes/:mesh/service-insights/:name",{mesh:s.params.mesh,name:s.params.service})},{default:a(({data:t})=>[o(V,{title:r("services.internal-service.about.title"),created:t.creationTime,modified:t.modificationTime},{default:a(()=>{var e,u;return[o(C,{layout:"horizontal"},{title:a(()=>[n(i(r("http.api.property.status")),1)]),body:a(()=>[o(w,{status:t.status},null,8,["status"])]),_:2},1024),l[2]||(l[2]=n()),o(C,{layout:"horizontal"},{title:a(()=>[n(i(r("http.api.property.address")),1)]),body:a(()=>[t.addressPort?(p(),f(x,{key:0,variant:"badge",format:"default",text:t.addressPort},null,8,["text"])):(p(),d(y,{key:1},[n(i(r("common.detail.none")),1)],64))]),_:2},1024),l[3]||(l[3]=n()),o(T,{layout:"horizontal",online:((e=t.dataplanes)==null?void 0:e.online)??0,total:((u=t.dataplanes)==null?void 0:u.total)??0},{title:a(()=>[n(i(r("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])]}),_:2},1032,["title","created","modified"])]),_:2},1032,["src"]),l[14]||(l[14]=n()),_("div",null,[_("h3",null,i(r("services.detail.data_plane_proxies")),1),l[13]||(l[13]=n()),o(B,{class:"mt-4"},{default:a(()=>[_("search",null,[o(G,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...z("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:t=>s.update({...Object.fromEntries(t.entries())})},null,8,["query","fields","onChange"])]),l[12]||(l[12]=n()),o(v,{src:k(b(q),"/meshes/:mesh/dataplanes/for/service-insight/:service",{mesh:s.params.mesh,service:s.params.service},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:a(({data:t})=>[o(D,{type:"data-planes",items:(t==null?void 0:t.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:t==null?void 0:t.total,onChange:s.update},{default:a(()=>[o(L,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.namespace"),label:"Namespace",key:"namespace"},...z("use zones")?[{...c.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...c.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":e=>e.name===s.params.proxy,onResize:c.set},{name:a(({row:e})=>[o(h,{"data-action":"",class:"name-link",to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,proxy:e.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:a(()=>[n(i(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(i(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(p(),f(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(i(e.zone),1)]),_:2},1032,["to"])):(p(),d(y,{key:1},[n(i(r("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var u;return[(u=e.dataplaneInsight.mTLS)!=null&&u.certificateExpirationTime?(p(),d(y,{key:0},[n(i(r("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(p(),d(y,{key:1},[n(i(r("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(w,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(p(),f(S,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[_("ul",null,[e.warnings.length>0?(p(),d("li",M,i(r("data-planes.components.data-plane-list.version_mismatch")),1)):g("",!0),l[4]||(l[4]=n()),e.isCertExpired?(p(),d("li",O,i(r("data-planes.components.data-plane-list.cert_expired")),1)):g("",!0)])]),_:2},1024)):(p(),d(y,{key:1},[n(i(r("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(X,null,{default:a(()=>[o(h,{to:{name:"data-plane-detail-view",params:{proxy:e.id}}},{default:a(()=>[n(i(r("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l[11]||(l[11]=n()),o(A,null,{default:a(({Component:e})=>[s.child()?(p(),f($,{key:0,onClose:u=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:a(()=>[typeof t<"u"?(p(),f(F(e),{key:0,items:t.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])):g("",!0)]),_:2},1024)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),K=P(U,[["__scopeId","data-v-3e096bd3"]]);export{K as default};
