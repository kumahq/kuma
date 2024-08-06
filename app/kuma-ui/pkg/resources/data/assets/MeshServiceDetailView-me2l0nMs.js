import{d as E,r as m,o as t,m as p,w as e,b as i,k as f,U as k,e as n,T as I,t as l,c as r,F as u,G as v,p as g,l as N,ay as K,A as X,S as q,E as F,q as $}from"./index-Is4zmHdk.js";import{F as G}from"./FilterBar-CdsH_AZN.js";import{S as M}from"./SummaryView-QbTr0JVE.js";const O={class:"stack"},j={class:"columns"},J={key:0},U={key:1},W=E({__name:"MeshServiceDetailView",props:{data:{}},setup(V){const _=V;return(y,Z)=>{const h=m("KTruncate"),b=m("KBadge"),w=m("KCard"),C=m("RouterLink"),S=m("XIcon"),P=m("XAction"),T=m("XActionGroup"),A=m("RouterView"),D=m("DataLoader"),L=m("AppView"),R=m("RouteView");return t(),p(R,{name:"mesh-service-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({can:x,route:o,t:c,uri:B,me:d})=>[i(L,null,{default:e(()=>[f("div",O,[i(w,null,{default:e(()=>[f("div",j,[_.data.status.addresses.length>0?(t(),p(k,{key:0},{title:e(()=>[n(`
                Addresses
              `)]),body:e(()=>[_.data.status.addresses.length===1?(t(),p(I,{key:0,text:_.data.status.addresses[0].hostname},{default:e(()=>[n(l(_.data.status.addresses[0].hostname),1)]),_:1},8,["text"])):(t(),p(h,{key:1},{default:e(()=>[(t(!0),r(u,null,v(_.data.status.addresses,s=>(t(),r("span",{key:s.hostname},l(s.hostname),1))),128))]),_:1}))]),_:1})):g("",!0),n(),i(k,null,{title:e(()=>[n(`
                Ports
              `)]),body:e(()=>[i(h,null,{default:e(()=>[(t(!0),r(u,null,v(y.data.spec.ports,s=>(t(),p(b,{key:s.port,appearance:"info"},{default:e(()=>[n(l(s.port)+":"+l(s.targetPort)+"/"+l(s.appProtocol),1)]),_:2},1024))),128))]),_:1})]),_:1}),n(),i(k,null,{title:e(()=>[n(`
                Dataplane Tags
              `)]),body:e(()=>[i(h,null,{default:e(()=>[(t(!0),r(u,null,v(y.data.spec.selector.dataplaneTags,(s,a)=>(t(),p(b,{key:`${a}:${s}`,appearance:"info"},{default:e(()=>[n(l(a)+":"+l(s),1)]),_:2},1024))),128))]),_:1})]),_:1}),n(),y.data.status.vips.length>0?(t(),p(k,{key:1,class:"ip"},{title:e(()=>[n(`
                VIPs
              `)]),body:e(()=>[i(h,null,{default:e(()=>[(t(!0),r(u,null,v(y.data.status.vips,s=>(t(),r("span",{key:s.ip},l(s.ip),1))),128))]),_:1})]),_:1})):g("",!0)])]),_:1}),n(),f("div",null,[f("h3",null,l(c("services.detail.data_plane_proxies")),1),n(),i(w,{class:"mt-4"},{default:e(()=>[i(D,{src:B(N(K),"/meshes/:mesh/dataplanes/for/mesh-service/:tags",{mesh:o.params.mesh,tags:JSON.stringify(_.data.spec.selector.dataplaneTags)},{page:o.params.page,size:o.params.size,search:o.params.s})},{loadable:e(({data:s})=>[i(X,{class:"data-plane-collection","data-testid":"data-plane-collection","page-number":o.params.page,"page-size":o.params.size,headers:[{...d.get("headers.name"),label:"Name",key:"name"},{...d.get("headers.namespace"),label:"Namespace",key:"namespace"},...x("use zones")?[{...d.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...d.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...d.get("headers.status"),label:"Status",key:"status"},{...d.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...d.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,total:s==null?void 0:s.total,"is-selected-row":a=>a.name===o.params.dataPlane,"summary-route-name":"service-data-plane-summary-view","empty-state-message":c("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":c("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":c("common.documentation"),onChange:o.update,onResize:d.set},{toolbar:e(()=>[i(G,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:o.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"},...x("use zones")&&{zone:{description:"filter by “kuma.io/zone” value"}}},onChange:a=>o.update({...Object.fromEntries(a.entries())})},null,8,["query","fields","onChange"])]),name:e(({row:a})=>[i(C,{class:"name-link",to:{name:"mesh-service-data-plane-summary-view",params:{mesh:a.mesh,dataPlane:a.id},query:{page:o.params.page,size:o.params.size,s:o.params.s}}},{default:e(()=>[n(l(a.name),1)]),_:2},1032,["to"])]),namespace:e(({row:a})=>[n(l(a.namespace),1)]),zone:e(({row:a})=>[a.zone?(t(),p(C,{key:0,to:{name:"zone-cp-detail-view",params:{zone:a.zone}}},{default:e(()=>[n(l(a.zone),1)]),_:2},1032,["to"])):(t(),r(u,{key:1},[n(l(c("common.collection.none")),1)],64))]),certificate:e(({row:a})=>{var z;return[(z=a.dataplaneInsight.mTLS)!=null&&z.certificateExpirationTime?(t(),r(u,{key:0},[n(l(c("common.formats.datetime",{value:Date.parse(a.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(t(),r(u,{key:1},[n(l(c("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:e(({row:a})=>[i(q,{status:a.status},null,8,["status"])]),warnings:e(({row:a})=>[a.isCertExpired||a.warnings.length>0?(t(),p(S,{key:0,class:"mr-1",name:"warning"},{default:e(()=>[f("ul",null,[a.warnings.length>0?(t(),r("li",J,l(c("data-planes.components.data-plane-list.version_mismatch")),1)):g("",!0),n(),a.isCertExpired?(t(),r("li",U,l(c("data-planes.components.data-plane-list.cert_expired")),1)):g("",!0)])]),_:2},1024)):(t(),r(u,{key:1},[n(l(c("common.collection.none")),1)],64))]),actions:e(({row:a})=>[i(T,null,{default:e(()=>[i(P,{to:{name:"data-plane-detail-view",params:{dataPlane:a.id}}},{default:e(()=>[n(l(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["page-number","page-size","headers","items","total","is-selected-row","empty-state-message","empty-state-cta-to","empty-state-cta-text","onChange","onResize"]),n(),o.params.dataPlane?(t(),p(A,{key:0},{default:e(a=>[i(M,{onClose:z=>o.replace({name:o.name,params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size,s:o.params.s}})},{default:e(()=>[typeof s<"u"?(t(),p(F(a.Component),{key:0,items:s.items},null,8,["items"])):g("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):g("",!0)]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),ee=$(W,[["__scopeId","data-v-89572731"]]);export{ee as default};
