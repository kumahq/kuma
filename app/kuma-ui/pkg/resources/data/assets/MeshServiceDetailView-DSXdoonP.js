import{d as N,r as c,o as i,m as p,w as e,b as o,k as _,Z as h,e as t,t as l,p as g,c as u,L as f,M as S,l as I,a9 as X,A as q,S as F,E as M,q as Z}from"./index-CMjLgvOo.js";import{F as $}from"./FilterBar-B7m2qAXa.js";import{S as G}from"./SummaryView-C546ionl.js";const O={class:"stack"},j={class:"columns"},J={key:0},W={key:1},H=N({__name:"MeshServiceDetailView",props:{data:{}},setup(V){const r=V;return(x,Q)=>{const k=c("KBadge"),v=c("XAction"),A=c("KumaPort"),b=c("KTruncate"),w=c("KCard"),C=c("RouterLink"),D=c("XIcon"),L=c("XActionGroup"),P=c("RouterView"),R=c("DataCollection"),T=c("DataLoader"),B=c("AppView"),E=c("RouteView");return i(),p(E,{name:"mesh-service-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({can:y,route:s,t:m,uri:K,me:d})=>[o(B,null,{default:e(()=>[_("div",O,[o(w,null,{default:e(()=>[_("div",j,[o(h,null,{title:e(()=>[t(`
                State
              `)]),body:e(()=>[o(k,{appearance:r.data.spec.state==="Available"?"success":"danger"},{default:e(()=>[t(l(r.data.spec.state),1)]),_:1},8,["appearance"])]),_:1}),t(),r.data.namespace.length>0?(i(),p(h,{key:0},{title:e(()=>[t(`
                Namespace
              `)]),body:e(()=>[t(l(r.data.namespace),1)]),_:1})):g("",!0),t(),y("use zones")&&r.data.zone?(i(),p(h,{key:1},{title:e(()=>[t(`
                Zone
              `)]),body:e(()=>[o(v,{to:{name:"zone-cp-detail-view",params:{zone:r.data.zone}}},{default:e(()=>[t(l(r.data.zone),1)]),_:1},8,["to"])]),_:1})):g("",!0),t(),o(h,null,{title:e(()=>[t(`
                Ports
              `)]),body:e(()=>[o(b,null,{default:e(()=>[(i(!0),u(f,null,S(r.data.spec.ports,n=>(i(),p(A,{key:n.port,port:{...n,targetPort:void 0}},null,8,["port"]))),128))]),_:1})]),_:1}),t(),o(h,null,{title:e(()=>[t(`
                Selector
              `)]),body:e(()=>[o(b,null,{default:e(()=>[(i(!0),u(f,null,S(x.data.spec.selector.dataplaneTags,(n,a)=>(i(),p(k,{key:`${a}:${n}`,appearance:"info"},{default:e(()=>[t(l(a)+":"+l(n),1)]),_:2},1024))),128))]),_:1})]),_:1})])]),_:2},1024),t(),_("div",null,[_("h3",null,l(m("services.detail.data_plane_proxies")),1),t(),o(w,{class:"mt-4"},{default:e(()=>[_("search",null,[o($,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:s.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"}},onChange:n=>s.update({...Object.fromEntries(n.entries())})},null,8,["query","onChange"])]),t(),o(T,{src:K(I(X),"/meshes/:mesh/dataplanes/for/mesh-service/:tags",{mesh:s.params.mesh,tags:JSON.stringify({...y("use zones")&&r.data.zone?{"kuma.io/zone":r.data.zone}:{},...r.data.spec.selector.dataplaneTags})},{page:s.params.page,size:s.params.size,search:s.params.s})},{loadable:e(({data:n})=>[o(R,{type:"data-planes",items:(n==null?void 0:n.items)??[void 0],page:s.params.page,"page-size":s.params.size,total:n==null?void 0:n.total,onChange:s.update},{default:e(()=>[o(q,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...d.get("headers.name"),label:"Name",key:"name"},{...d.get("headers.namespace"),label:"Namespace",key:"namespace"},...y("use zones")?[{...d.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...d.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...d.get("headers.status"),label:"Status",key:"status"},{...d.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...d.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:n==null?void 0:n.items,"is-selected-row":a=>a.name===s.params.dataPlane,onResize:d.set},{name:e(({row:a})=>[o(C,{class:"name-link",to:{name:"mesh-service-data-plane-summary-view",params:{mesh:a.mesh,dataPlane:a.id},query:{page:s.params.page,size:s.params.size,s:s.params.s}}},{default:e(()=>[t(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[t(l(a.namespace),1)]),zone:e(({row:a})=>[a.zone?(i(),p(C,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:e(()=>[t(l(a.zone),1)]),_:2},1032,["to"])):(i(),u(f,{key:1},[t(l(m("common.collection.none")),1)],64))]),certificate:e(({row:a})=>{var z;return[(z=a.dataplaneInsight.mTLS)!=null&&z.certificateExpirationTime?(i(),u(f,{key:0},[t(l(m("common.formats.datetime",{value:Date.parse(a.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(i(),u(f,{key:1},[t(l(m("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:e(({row:a})=>[o(F,{status:a.status},null,8,["status"])]),warnings:e(({row:a})=>[a.isCertExpired||a.warnings.length>0?(i(),p(D,{key:0,class:"mr-1",name:"warning"},{default:e(()=>[_("ul",null,[a.warnings.length>0?(i(),u("li",J,l(m("data-planes.components.data-plane-list.version_mismatch")),1)):g("",!0),t(),a.isCertExpired?(i(),u("li",W,l(m("data-planes.components.data-plane-list.cert_expired")),1)):g("",!0)])]),_:2},1024)):(i(),u(f,{key:1},[t(l(m("common.collection.none")),1)],64))]),actions:e(({row:a})=>[o(L,null,{default:e(()=>[o(v,{to:{name:"data-plane-detail-view",params:{dataPlane:a.id}}},{default:e(()=>[t(l(m("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),t(),s.params.dataPlane?(i(),p(P,{key:0},{default:e(a=>[o(G,{onClose:z=>s.replace({name:s.name,params:{mesh:s.params.mesh},query:{page:s.params.page,size:s.params.size,s:s.params.s}})},{default:e(()=>[typeof n<"u"?(i(),p(M(a.Component),{key:0,items:n.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),ae=Z(H,[["__scopeId","data-v-8c1630b3"]]);export{ae as default};
