import{d as $,r,o,q as u,w as s,b as n,m as k,e as i,B as I,t as l,c as d,M as _,S as N,s as g,I as E,_ as L}from"./index-CvmQ8qmc.js";import{F as R}from"./FilterBar-1idCMK4g.js";import{S as q}from"./SummaryView-CCQSXOe-.js";const T={class:"stack"},F={key:0},G={key:1},j=$({__name:"BuiltinGatewayDataplanesView",setup(M){return(O,p)=>{const y=r("XAction"),b=r("XIcon"),v=r("XActionGroup"),w=r("RouterView"),C=r("DataCollection"),x=r("DataLoader"),V=r("XCard"),S=r("DataSource"),B=r("AppView"),A=r("RouteView");return o(),u(A,{name:"builtin-gateway-dataplanes-view",params:{mesh:"",gateway:"",listener:"",page:1,size:Number,s:"",proxy:""}},{default:s(({can:z,route:a,t:c,me:m})=>[n(B,null,{default:s(()=>[n(S,{src:`/meshes/${a.params.mesh}/mesh-gateways/${a.params.gateway}`},{default:s(({data:f,error:X})=>[k("div",T,[n(V,null,{default:s(()=>[k("search",null,[n(R,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:a.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},service:{description:"filter by “kuma.io/service” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...z("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:t=>a.update({page:1,...Object.fromEntries(t.entries())})},null,8,["query","fields","onChange"])]),p[8]||(p[8]=i()),n(x,{src:f===void 0?"":`/meshes/${a.params.mesh}/dataplanes/for/service-insight/${f.selectors[0].match["kuma.io/service"]}?page=${a.params.page}&size=${a.params.size}&search=${a.params.s}`,data:[f],errors:[X]},{loadable:s(({data:t})=>[n(C,{type:"data-planes",items:(t==null?void 0:t.items)??[void 0],total:t==null?void 0:t.total,page:a.params.page,"page-size":a.params.size,onChange:a.update},{default:s(()=>[n(I,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.namespace"),label:"Namespace",key:"namespace"},...z("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...m.get("headers.status"),label:"Status",key:"status"},{...m.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,"is-selected-row":e=>e.name===a.params.proxy,onResize:m.set},{namespace:s(({row:e})=>[i(l(e.namespace),1)]),name:s(({row:e})=>[n(y,{"data-action":"",class:"name-link",title:e.name,to:{name:"builtin-gateway-data-plane-summary-view",params:{mesh:e.mesh,proxy:e.id},query:{page:a.params.page,size:a.params.size,s:a.params.s}}},{default:s(()=>[i(l(e.name),1)]),_:2},1032,["title","to"])]),zone:s(({row:e})=>[e.zone?(o(),u(y,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[i(l(e.zone),1)]),_:2},1032,["to"])):(o(),d(_,{key:1},[i(l(c("common.collection.none")),1)],64))]),certificate:s(({row:e})=>{var h;return[(h=e.dataplaneInsight.mTLS)!=null&&h.certificateExpirationTime?(o(),d(_,{key:0},[i(l(c("common.formats.datetime",{value:Date.parse(e.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(o(),d(_,{key:1},[i(l(c("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:s(({row:e})=>[n(N,{status:e.status},null,8,["status"])]),warnings:s(({row:e})=>[e.isCertExpired||e.warnings.length>0?(o(),u(b,{key:0,class:"mr-1",name:"warning"},{default:s(()=>[k("ul",null,[e.warnings.length>0?(o(),d("li",F,l(c("data-planes.components.data-plane-list.version_mismatch")),1)):g("",!0),p[0]||(p[0]=i()),e.isCertExpired?(o(),d("li",G,l(c("data-planes.components.data-plane-list.cert_expired")),1)):g("",!0)])]),_:2},1024)):(o(),d(_,{key:1},[i(l(c("common.collection.none")),1)],64))]),actions:s(({row:e})=>[n(v,null,{default:s(()=>[n(y,{to:{name:"data-plane-detail-view",params:{proxy:e.id}}},{default:s(()=>[i(l(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),p[7]||(p[7]=i()),a.params.proxy?(o(),u(w,{key:0},{default:s(e=>[n(q,{onClose:h=>a.replace({name:a.name,params:{mesh:a.params.mesh},query:{page:a.params.page,size:a.params.size,s:a.params.s}})},{default:s(()=>[typeof t<"u"?(o(),u(E(e.Component),{key:0,items:t.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["items","total","page","page-size","onChange"])]),_:2},1032,["src","data","errors"])]),_:2},1024)])]),_:2},1032,["src"])]),_:2},1024)]),_:1})}}}),J=L(j,[["__scopeId","data-v-c8b6aba2"]]);export{J as default};
