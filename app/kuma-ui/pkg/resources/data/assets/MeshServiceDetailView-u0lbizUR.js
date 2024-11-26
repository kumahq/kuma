import{d as L,e as p,o as l,m as d,w as e,a as i,k as f,Q as z,b as s,t as r,p as _,c as g,J as y,K as S,l as R,ah as I,A as q,S as F,F as $,q as G}from"./index-CKcsX_-l.js";import{F as J}from"./FilterBar-DdKAp1jk.js";import{S as M}from"./SummaryView-BIwsKbzL.js";const O={class:"stack"},Z={class:"columns"},j={key:0},Q={key:1},W=L({__name:"MeshServiceDetailView",props:{data:{}},setup(V){const m=V;return(x,a)=>{const w=p("XBadge"),k=p("XAction"),A=p("KumaPort"),C=p("KTruncate"),h=p("KCard"),D=p("XIcon"),P=p("XActionGroup"),T=p("RouterView"),B=p("DataCollection"),N=p("DataLoader"),X=p("AppView"),E=p("RouteView");return l(),d(E,{name:"mesh-service-detail-view",params:{mesh:"",service:"",page:1,size:50,s:"",dataPlane:"",codeSearch:"",codeFilter:!1,codeRegExp:!1}},{default:e(({can:v,route:n,t:c,uri:K,me:u})=>[i(X,null,{default:e(()=>[f("div",O,[i(h,null,{default:e(()=>[f("div",Z,[i(z,null,{title:e(()=>a[0]||(a[0]=[s(`
                State
              `)])),body:e(()=>[i(w,{appearance:m.data.spec.state==="Available"?"success":"danger"},{default:e(()=>[s(r(m.data.spec.state),1)]),_:1},8,["appearance"])]),_:1}),a[10]||(a[10]=s()),m.data.namespace.length>0?(l(),d(z,{key:0},{title:e(()=>a[2]||(a[2]=[s(`
                Namespace
              `)])),body:e(()=>[s(r(m.data.namespace),1)]),_:1})):_("",!0),a[11]||(a[11]=s()),v("use zones")&&m.data.zone?(l(),d(z,{key:1},{title:e(()=>a[4]||(a[4]=[s(`
                Zone
              `)])),body:e(()=>[i(k,{to:{name:"zone-cp-detail-view",params:{zone:m.data.zone}}},{default:e(()=>[s(r(m.data.zone),1)]),_:1},8,["to"])]),_:1})):_("",!0),a[12]||(a[12]=s()),i(z,null,{title:e(()=>a[6]||(a[6]=[s(`
                Ports
              `)])),body:e(()=>[i(C,null,{default:e(()=>[(l(!0),g(y,null,S(m.data.spec.ports,o=>(l(),d(A,{key:o.port,port:{...o,targetPort:void 0}},null,8,["port"]))),128))]),_:1})]),_:1}),a[13]||(a[13]=s()),i(z,null,{title:e(()=>a[8]||(a[8]=[s(`
                Selector
              `)])),body:e(()=>[i(C,null,{default:e(()=>[(l(!0),g(y,null,S(x.data.spec.selector.dataplaneTags,(o,t)=>(l(),d(w,{key:`${t}:${o}`,appearance:"info"},{default:e(()=>[s(r(t)+":"+r(o),1)]),_:2},1024))),128))]),_:1})]),_:1})])]),_:2},1024),a[24]||(a[24]=s()),f("div",null,[f("h3",null,r(c("services.detail.data_plane_proxies")),1),a[23]||(a[23]=s()),i(h,{class:"mt-4"},{default:e(()=>[f("search",null,[i(J,{class:"data-plane-proxy-filter",placeholder:"name:dataplane-name",query:n.params.s,fields:{name:{description:"filter by name or parts of a name"},protocol:{description:"filter by “kuma.io/protocol” value"},tag:{description:"filter by tags (e.g. “tag: version:2”)"}},onChange:o=>n.update({...Object.fromEntries(o.entries())})},null,8,["query","onChange"])]),a[22]||(a[22]=s()),i(N,{src:K(R(I),"/meshes/:mesh/dataplanes/for/mesh-service/:tags",{mesh:n.params.mesh,tags:JSON.stringify({...v("use zones")&&m.data.zone?{"kuma.io/zone":m.data.zone}:{},...m.data.spec.selector.dataplaneTags})},{page:n.params.page,size:n.params.size,search:n.params.s})},{loadable:e(({data:o})=>[i(B,{type:"data-planes",items:(o==null?void 0:o.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:o==null?void 0:o.total,onChange:n.update},{default:e(()=>[i(q,{class:"data-plane-collection","data-testid":"data-plane-collection",headers:[{...u.get("headers.name"),label:"Name",key:"name"},{...u.get("headers.namespace"),label:"Namespace",key:"namespace"},...v("use zones")?[{...u.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...u.get("headers.certificate"),label:"Certificate Info",key:"certificate"},{...u.get("headers.status"),label:"Status",key:"status"},{...u.get("headers.warnings"),label:"Warnings",key:"warnings",hideLabel:!0},{...u.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,"is-selected-row":t=>t.name===n.params.dataPlane,onResize:u.set},{name:e(({row:t})=>[i(k,{class:"name-link",to:{name:"mesh-service-data-plane-summary-view",params:{mesh:t.mesh,dataPlane:t.id},query:{page:n.params.page,size:n.params.size,s:n.params.s}},"data-action":""},{default:e(()=>[s(r(t.name),1)]),_:2},1032,["to"])]),namespace:e(({row:t})=>[s(r(t.namespace),1)]),zone:e(({row:t})=>[t.zone?(l(),d(k,{key:0,to:{name:"zone-cp-detail-view",params:{zone:t.zone}}},{default:e(()=>[s(r(t.zone),1)]),_:2},1032,["to"])):(l(),g(y,{key:1},[s(r(c("common.collection.none")),1)],64))]),certificate:e(({row:t})=>{var b;return[(b=t.dataplaneInsight.mTLS)!=null&&b.certificateExpirationTime?(l(),g(y,{key:0},[s(r(c("common.formats.datetime",{value:Date.parse(t.dataplaneInsight.mTLS.certificateExpirationTime)})),1)],64)):(l(),g(y,{key:1},[s(r(c("data-planes.components.data-plane-list.certificate.none")),1)],64))]}),status:e(({row:t})=>[i(F,{status:t.status},null,8,["status"])]),warnings:e(({row:t})=>[t.isCertExpired||t.warnings.length>0?(l(),d(D,{key:0,class:"mr-1",name:"warning"},{default:e(()=>[f("ul",null,[t.warnings.length>0?(l(),g("li",j,r(c("data-planes.components.data-plane-list.version_mismatch")),1)):_("",!0),a[14]||(a[14]=s()),t.isCertExpired?(l(),g("li",Q,r(c("data-planes.components.data-plane-list.cert_expired")),1)):_("",!0)])]),_:2},1024)):(l(),g(y,{key:1},[s(r(c("common.collection.none")),1)],64))]),actions:e(({row:t})=>[i(P,null,{default:e(()=>[i(k,{to:{name:"data-plane-detail-view",params:{dataPlane:t.id}}},{default:e(()=>[s(r(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),a[21]||(a[21]=s()),n.params.dataPlane?(l(),d(T,{key:0},{default:e(t=>[i(M,{onClose:b=>n.replace({name:n.name,params:{mesh:n.params.mesh},query:{page:n.params.page,size:n.params.size,s:n.params.s}})},{default:e(()=>[typeof o<"u"?(l(),d($(t.Component),{key:0,items:o.items},null,8,["items"])):_("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):_("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),ee=G(W,[["__scopeId","data-v-b88b6d8e"]]);export{ee as default};
