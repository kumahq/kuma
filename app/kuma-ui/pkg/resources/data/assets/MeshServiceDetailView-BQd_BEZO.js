import{d as I,r as d,o as r,q as u,w as e,b as n,U as _,e as t,t as s,s as f,c,M as g,N as C,m as k,p as R,Y as q,B as F,S as M,I as O,_ as $}from"./index-Du84oSnm.js";import{F as j}from"./FilterBar-CKqqItEJ.js";import{S as G}from"./SummaryView-Cd8oe3uM.js";const K={key:0},J={key:1},U=I({__name:"MeshServiceDetailView",props:{data:{}},setup(w){const p=w;return(W,l)=>{const h=d("XBadge"),z=d("XAction"),X=d("KumaPort"),x=d("XAboutCard"),V=d("XIcon"),S=d("XActionGroup"),A=d("RouterView"),D=d("DataCollection"),T=d("DataLoader"),B=d("XCard"),L=d("XLayout"),N=d("AppView"),P=d("RouteView");return r(),u(P,{name:"mesh-service-detail-view",params:{mesh:"",service:"",page:1,size:Number,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({can:v,route:o,t:m,uri:E,me:y})=>[n(N,null,{default:e(()=>[n(L,{type:"stack"},{default:e(()=>[n(x,{title:m("services.mesh-service.about.title"),created:p.data.creationTime,modified:p.data.modificationTime},{default:e(()=>[n(_,{layout:"horizontal"},{title:e(()=>[t(s(m("http.api.property.state")),1)]),body:e(()=>[n(h,{appearance:p.data.spec.state==="Available"?"success":"danger"},{default:e(()=>[t(s(p.data.spec.state),1)]),_:1},8,["appearance"])]),_:2},1024),l[5]||(l[5]=t()),p.data.namespace.length>0?(r(),u(_,{key:0,layout:"horizontal"},{title:e(()=>[t(s(m("http.api.property.namespace")),1)]),body:e(()=>[n(h,{appearance:"decorative"},{default:e(()=>[t(s(p.data.namespace),1)]),_:1})]),_:2},1024)):f("",!0),l[6]||(l[6]=t()),v("use zones")&&p.data.zone?(r(),u(_,{key:1,layout:"horizontal"},{title:e(()=>[t(s(m("http.api.property.zone")),1)]),body:e(()=>[n(h,{appearance:"decorative"},{default:e(()=>[n(z,{to:{name:"zone-cp-detail-view",params:{zone:p.data.zone}}},{default:e(()=>[t(s(p.data.zone),1)]),_:1},8,["to"])]),_:1})]),_:2},1024)):f("",!0),l[7]||(l[7]=t()),n(_,{layout:"horizontal"},{title:e(()=>[t(s(m("http.api.property.ports")),1)]),body:e(()=>[p.data.spec.ports.length?(r(!0),c(g,{key:0},C(p.data.spec.ports,i=>(r(),u(X,{key:i.port,port:{...i,targetPort:void 0}},null,8,["port"]))),128)):(r(),c(g,{key:1},[t(s(m("common.detail.none")),1)],64))]),_:2},1024),l[8]||(l[8]=t()),n(_,{layout:"horizontal"},{title:e(()=>[t(s(m("http.api.property.selector")),1)]),body:e(()=>[Object.keys(p.data.spec.selector.dataplaneTags).length?(r(!0),c(g,{key:0},C(p.data.spec.selector.dataplaneTags,(i,a)=>(r(),u(h,{key:`${a}:${i}`,appearance:"info"},{default:e(()=>[t(s(a)+":"+s(i),1)]),_:2},1024))),128)):(r(),c(g,{key:1},[t(s(m("common.detail.none")),1)],64))]),_:2},1024)]),_:2},1032,["title","created","modified"]),l[19]||(l[19]=t()),k("div",null,[k("h3",null,s(m("services.detail.data_plane_proxies")),1),l[18]||(l[18]=t()),n(B,{class:"mt-4"},{default:e(()=>[k("search",null,[n(j,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:o.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"}},onChange:i=>o.update({...Object.fromEntries(i.entries())})},null,8,["query","onChange"])]),l[17]||(l[17]=t()),n(T,{src:E(R(q),"/meshes/:mesh/dataplanes/for/mesh-service/:tags",{mesh:o.params.mesh,tags:JSON.stringify({...v("use zones")&&p.data.zone?{"kuma.io/zone":p.data.zone}:{},...p.data.spec.selector.dataplaneTags})},{page:o.params.page,size:o.params.size,search:o.params.s})},{loadable:e(({data:i})=>[n(D,{type:"data-planes",items:(i==null?void 0:i.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:i==null?void 0:i.total,onChange:o.update},{default:e(()=>[n(F,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...y.get("headers.name"),label:"Name",key:"name"},{...y.get("headers.namespace"),label:"Namespace",key:"namespace"},...v("use zones")?[{...y.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...y.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...y.get("headers.status"),label:"Status",key:"status"},{...y.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...y.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:i==null?void 0:i.items,"is-selected-row":a=>a.name===o.params.dataPlane,onResize:y.set},{name:e(({row:a})=>[n(z,{class:"name-link",to:{name:"mesh-service-data-plane-summary-view",params:{mesh:a.mesh,dataPlane:a.id},query:{page:o.params.page,size:o.params.size,s:o.params.s}},"data-action":""},{default:e(()=>[t(s(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[t(s(a.namespace),1)]),zone:e(({row:a})=>[a.zone?(r(),u(z,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:e(()=>[t(s(a.zone),1)]),_:2},1032,["to"])):(r(),c(g,{key:1},[t(s(m("common.collection.none")),1)],64))]),certificate:e(({row:a})=>{var b;return[(b=a.dataplaneInsight.mTLS)!=null&&b.certificateExpirationTime?(r(),c(g,{key:0},[t(s(m("common.formats.datetime",{value:Date.parse(a.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(r(),c(g,{key:1},[t(s(m("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:e(({row:a})=>[n(M,{status:a.status},null,8,["status"])]),warnings:e(({row:a})=>[a.isCertExpired||a.warnings.length>0?(r(),u(V,{key:0,class:"mr-1",name:"warning"},{default:e(()=>[k("ul",null,[a.warnings.length>0?(r(),c("li",K,s(m("data-planes.components.data-plane-list.version_mismatch")),1)):f("",!0),l[9]||(l[9]=t()),a.isCertExpired?(r(),c("li",J,s(m("data-planes.components.data-plane-list.cert_expired")),1)):f("",!0)])]),_:2},1024)):(r(),c(g,{key:1},[t(s(m("common.collection.none")),1)],64))]),actions:e(({row:a})=>[n(S,null,{default:e(()=>[n(z,{to:{name:"data-plane-detail-view",params:{dataPlane:a.id}}},{default:e(()=>[t(s(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l[16]||(l[16]=t()),o.params.dataPlane?(r(),u(A,{key:0},{default:e(a=>[n(G,{onClose:b=>o.replace({name:o.name,params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size,s:o.params.s}})},{default:e(()=>[typeof i<"u"?(r(),u(O(a.Component),{key:0,items:i.items},null,8,["items"])):f("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):f("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1024)]),_:1})}}}),Q=$(U,[["__scopeId","data-v-4d721d5b"]]);export{Q as default};
