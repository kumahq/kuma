import{d as N,i as n,o as l,a as c,w as s,j as t,g as B,k as m,A as f,az as I,a3 as X,t as i,b as p,H as _,e as h,J as v,K as $,r as P,_ as q}from"./index-Dh8cIC_Y.js";import{e as E}from"./kong-icons.es249-B6BLouOf.js";import{A as M}from"./AppCollection-CXZmK54L.js";import{S as Z}from"./SummaryView-1_MRdDY6.js";import"./kong-icons.es245-DFixHsMJ.js";const j=N({__name:"MeshServiceListView",setup(F){return(H,J)=>{const y=n("RouteTitle"),b=n("XTeleportTemplate"),k=n("XAction"),C=n("RouterLink"),u=n("KTruncate"),w=n("KBadge"),V=n("RouterView"),T=n("DataCollection"),D=n("DataLoader"),S=n("KCard"),A=n("AppView"),K=n("RouteView"),L=n("DataSource");return l(),c(L,{src:"/me"},{default:s(({data:g})=>[g?(l(),c(K,{key:0,name:"mesh-service-list-view",params:{page:1,size:g.pageSize,mesh:"",service:""}},{default:s(({route:o,t:d,can:R,uri:x})=>[t(A,null,{title:s(()=>[t(b,{to:{name:"service-list-tabs-view-title"}},{default:s(()=>[B("h2",null,[t(y,{title:d("services.routes.mesh-service-list-view.title")},null,8,["title"])])]),_:2},1024)]),default:s(()=>[m(),t(S,null,{default:s(()=>[t(D,{src:x(f(I),"/meshes/:mesh/mesh-services",{mesh:o.params.mesh},{page:o.params.page,size:o.params.size})},{loadable:s(({data:a})=>[t(T,{type:"services",items:(a==null?void 0:a.items)??[void 0]},{default:s(()=>[t(M,{"data-testid":"service-collection",headers:[{label:"Name",key:"name"},{label:"Namespace",key:"namespace"},...R("use zones")?[{label:"Zone",key:"zone"}]:[],{label:"Addresses",key:"addresses"},{label:"Ports",key:"ports"},{label:"Tags",key:"tags"},{label:"Details",key:"details",hideLabel:!0}],"page-number":o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,"is-selected-row":e=>e.name===o.params.service,onChange:o.update},{name:s(({row:e})=>[t(X,{text:e.name},{default:s(()=>[t(k,{"data-action":"",to:{name:"mesh-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:o.params.page,size:o.params.size}}},{default:s(()=>[m(i(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[m(i(e.namespace),1)]),zone:s(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(l(),p(_,{key:0},[e.labels["kuma.io/zone"]?(l(),c(C,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:s(()=>[m(i(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):h("",!0)],64)):(l(),p(_,{key:1},[m(i(d("common.detail.none")),1)],64))]),addresses:s(({row:e})=>[t(u,null,{default:s(()=>[(l(!0),p(_,null,v(e.status.addresses,r=>(l(),p("span",{key:r.hostname},i(r.hostname),1))),128))]),_:2},1024)]),ports:s(({row:e})=>[t(u,null,{default:s(()=>[(l(!0),p(_,null,v(e.spec.ports,r=>(l(),c(w,{key:r.port,appearance:"info"},{default:s(()=>[m(i(r.port)+":"+i(r.targetPort)+"/"+i(r.appProtocol),1)]),_:2},1024))),128))]),_:2},1024)]),tags:s(({row:e})=>[t(u,null,{default:s(()=>[(l(!0),p(_,null,v(e.spec.selector.dataplaneTags,(r,z)=>(l(),c(w,{key:`${z}:${r}`,appearance:"info"},{default:s(()=>[m(i(z)+":"+i(r),1)]),_:2},1024))),128))]),_:2},1024)]),details:s(({row:e})=>[t(k,{class:"details-link","data-testid":"details-link",to:{name:"mesh-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[m(i(d("common.collection.details_link"))+" ",1),t(f(E),{decorative:"",size:f($)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange"]),m(),a!=null&&a.items&&o.params.service?(l(),c(V,{key:0},{default:s(e=>[t(Z,{onClose:r=>o.replace({name:"mesh-service-list-view",params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size}})},{default:s(()=>[(l(),c(P(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):h("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1024)]),_:2},1032,["params"])):h("",!0)]),_:1})}}}),Y=q(j,[["__scopeId","data-v-a63e2c2c"]]);export{Y as default};
