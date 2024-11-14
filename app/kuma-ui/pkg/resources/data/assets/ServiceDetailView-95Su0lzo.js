import{d as E,e as m,o as r,m as _,w as a,a as o,k as f,l as C,ap as B,P as b,b as n,t as i,S as x,$ as I,c as d,H as g,R as L,ag as N,A as X,p as y,E as q,q as T}from"./index-B_icS-nL.js";import{F}from"./FilterBar-xr9ylAGx.js";import{S as $}from"./SummaryView-DoPm-AWH.js";const G={class:"stack"},K={class:"columns"},j={key:0},H={key:1},O=E({__name:"ServiceDetailView",setup(W){return(Z,l)=>{const v=m("DataLoader"),k=m("KCard"),h=m("XAction"),V=m("XIcon"),S=m("XActionGroup"),D=m("RouterView"),A=m("DataCollection"),P=m("AppView"),R=m("RouteView");return r(),_(R,{name:"service-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({can:z,route:s,t:p,me:c,uri:w})=>[o(P,null,{default:a(()=>[f("div",G,[o(k,null,{default:a(()=>[o(v,{src:w(C(B),"/meshes/:mesh/service-insights/:name",{mesh:s.params.mesh,name:s.params.service})},{default:a(({data:t})=>{var e,u;return[f("div",K,[o(b,null,{title:a(()=>[n(i(p("http.api.property.status")),1)]),body:a(()=>[o(x,{status:t.status},null,8,["status"])]),_:2},1024),l[2]||(l[2]=n()),o(b,null,{title:a(()=>[n(i(p("http.api.property.address")),1)]),body:a(()=>[t.addressPort?(r(),_(I,{key:0,text:t.addressPort},null,8,["text"])):(r(),d(g,{key:1},[n(i(p("common.detail.none")),1)],64))]),_:2},1024),l[3]||(l[3]=n()),o(L,{online:((e=t.dataplanes)==null?void 0:e.online)??0,total:((u=t.dataplanes)==null?void 0:u.total)??0},{title:a(()=>[n(i(p("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1032,["src"])]),_:2},1024),l[14]||(l[14]=n()),f("div",null,[f("h3",null,i(p("services.detail.data_plane_proxies")),1),l[13]||(l[13]=n()),o(k,{class:"mt-4"},{default:a(()=>[f("search",null,[o(F,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...z("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:t=>s.update({...Object.fromEntries(t.entries())})},null,8,["query","fields","onChange"])]),l[12]||(l[12]=n()),o(v,{src:w(C(N),"/meshes/:mesh/dataplanes/for/service-insight/:service",{mesh:s.params.mesh,service:s.params.service},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:a(({data:t})=>[o(A,{type:"data-planes",items:(t==null?void 0:t.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:t==null?void 0:t.total,onChange:s.update},{default:a(()=>[o(X,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.namespace"),label:"Namespace",key:"namespace"},...z("use zones")?[{...c.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...c.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":e=>e.name===s.params.dataPlane,onResize:c.set},{name:a(({row:e})=>[o(h,{"data-action":"",class:"name-link",to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:a(()=>[n(i(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[n(i(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(r(),_(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[n(i(e.zone),1)]),_:2},1032,["to"])):(r(),d(g,{key:1},[n(i(p("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var u;return[(u=e.dataplaneInsight.mTLS)!=null&&u.certificateExpirationTime?(r(),d(g,{key:0},[n(i(p("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(r(),d(g,{key:1},[n(i(p("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(x,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(r(),_(V,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[f("ul",null,[e.warnings.length>0?(r(),d("li",j,i(p("data-planes.components.data-plane-list.version_mismatch")),1)):y("",!0),l[4]||(l[4]=n()),e.isCertExpired?(r(),d("li",H,i(p("data-planes.components.data-plane-list.cert_expired")),1)):y("",!0)])]),_:2},1024)):(r(),d(g,{key:1},[n(i(p("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(S,null,{default:a(()=>[o(h,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[n(i(p("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l[11]||(l[11]=n()),o(D,null,{default:a(({Component:e})=>[s.child()?(r(),_($,{key:0,onClose:u=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:a(()=>[typeof t<"u"?(r(),_(q(e),{key:0,items:t.items},null,8,["items"])):y("",!0)]),_:2},1032,["onClose"])):y("",!0)]),_:2},1024)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),U=T(O,[["__scopeId","data-v-6253b442"]]);export{U as default};
