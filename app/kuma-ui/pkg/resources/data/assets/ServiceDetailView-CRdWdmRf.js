import{d as P,e as r,o as l,m as _,w as a,a as o,k as u,l as w,ap as R,X as C,b as t,t as i,S as b,$ as X,c as m,H as h,Y as B,a4 as I,A as L,p as f,E as N,q}from"./index-BGYhp_E8.js";import{F as T}from"./FilterBar-DVMeLfLQ.js";import{S as F}from"./SummaryView-DVexZK_i.js";const $={class:"stack"},G={class:"columns"},K={key:0},j={key:1},H=P({__name:"ServiceDetailView",setup(O){return(W,Y)=>{const y=r("DataLoader"),v=r("KCard"),g=r("XAction"),x=r("XIcon"),V=r("XActionGroup"),S=r("RouterView"),D=r("DataCollection"),A=r("AppView"),E=r("RouteView");return l(),_(E,{name:"service-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:a(({can:k,route:s,t:c,me:p,uri:z})=>[o(A,null,{default:a(()=>[u("div",$,[o(v,null,{default:a(()=>[o(y,{src:z(w(R),"/meshes/:mesh/service-insights/:name",{mesh:s.params.mesh,name:s.params.service})},{default:a(({data:n})=>{var e,d;return[u("div",G,[o(C,null,{title:a(()=>[t(i(c("http.api.property.status")),1)]),body:a(()=>[o(b,{status:n.status},null,8,["status"])]),_:2},1024),t(),o(C,null,{title:a(()=>[t(i(c("http.api.property.address")),1)]),body:a(()=>[n.addressPort?(l(),_(X,{key:0,text:n.addressPort},null,8,["text"])):(l(),m(h,{key:1},[t(i(c("common.detail.none")),1)],64))]),_:2},1024),t(),o(B,{online:((e=n.dataplanes)==null?void 0:e.online)??0,total:((d=n.dataplanes)==null?void 0:d.total)??0},{title:a(()=>[t(i(c("http.api.property.dataPlaneProxies")),1)]),_:2},1032,["online","total"])])]}),_:2},1032,["src"])]),_:2},1024),t(),u("div",null,[u("h3",null,i(c("services.detail.data_plane_proxies")),1),t(),o(v,{class:"mt-4"},{default:a(()=>[o(y,{src:z(w(I),"/meshes/:mesh/dataplanes/for/service-insight/:service",{mesh:s.params.mesh,service:s.params.service},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:a(({data:n})=>[o(D,{type:"data-planes",items:(n==null?void 0:n.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:n==null?void 0:n.total,onChange:s.update},{default:a(()=>[o(L,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...p.get("headers.name"),label:"Name",key:"name"},{...p.get("headers.namespace"),label:"Namespace",key:"namespace"},...k("use zones")?[{...p.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...p.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...p.get("headers.status"),label:"Status",key:"status"},{...p.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...p.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:n==null?void 0:n.items,"is-selected-row":e=>e.name===s.params.dataPlane,onResize:p.set},{toolbar:a(()=>[o(T,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...k("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:e=>s.update({...Object.fromEntries(e.entries())})},null,8,["query","fields","onChange"])]),name:a(({row:e})=>[o(g,{"data-action":"",class:"name-link",to:{name:"service-data-plane-summary-view",params:{mesh:e.mesh,dataPlane:e.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:a(()=>[t(i(e.name),1)]),_:2},1032,["to"])]),namespace:a(({row:e})=>[t(i(e.namespace),1)]),zone:a(({row:e})=>[e.zone?(l(),_(g,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:a(()=>[t(i(e.zone),1)]),_:2},1032,["to"])):(l(),m(h,{key:1},[t(i(c("common.collection.none")),1)],64))]),certificate:a(({row:e})=>{var d;return[(d=e.dataplaneInsight.mTLS)!=null&&d.certificateExpirationTime?(l(),m(h,{key:0},[t(i(c("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(l(),m(h,{key:1},[t(i(c("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:a(({row:e})=>[o(b,{status:e.status},null,8,["status"])]),warnings:a(({row:e})=>[e.isCertExpired||e.warnings.length>0?(l(),_(x,{key:0,class:"mr-1",name:"warning"},{default:a(()=>[u("ul",null,[e.warnings.length>0?(l(),m("li",K,i(c("data-planes.components.data-plane-list.version_mismatch")),1)):f("",!0),t(),e.isCertExpired?(l(),m("li",j,i(c("data-planes.components.data-plane-list.cert_expired")),1)):f("",!0)])]),_:2},1024)):(l(),m(h,{key:1},[t(i(c("common.collection.none")),1)],64))]),actions:a(({row:e})=>[o(V,null,{default:a(()=>[o(g,{to:{name:"data-plane-detail-view",params:{dataPlane:e.id}}},{default:a(()=>[t(i(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),t(),o(S,null,{default:a(({Component:e})=>[s.child()?(l(),_(F,{key:0,onClose:d=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:a(()=>[typeof n<"u"?(l(),_(N(e),{key:0,items:n.items},null,8,["items"])):f("",!0)]),_:2},1032,["onClose"])):f("",!0)]),_:2},1024)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),Q=q(H,[["__scopeId","data-v-c52f65c0"]]);export{Q as default};
