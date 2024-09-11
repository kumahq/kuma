import{d as K,r,o as t,m as p,w as s,b as n,e as m,l as B,ax as N,A as S,a2 as T,t as i,c as u,L as _,p as w,M as f,E as M}from"./index-DW4c3gZM.js";import{S as X}from"./SummaryView-BNxXmvI7.js";const Z=K({__name:"MeshMultiZoneServiceListView",setup($){return(P,q)=>{const b=r("RouteTitle"),d=r("XAction"),z=r("KTruncate"),g=r("KBadge"),k=r("XActionGroup"),y=r("RouterView"),C=r("DataCollection"),V=r("DataLoader"),A=r("KCard"),L=r("AppView"),R=r("RouteView");return t(),p(R,{name:"mesh-multi-zone-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:s(({route:o,t:h,can:x,uri:D,me:c})=>[n(b,{render:!1,title:h("services.routes.mesh-multi-zone-service-list-view.title")},null,8,["title"]),m(),n(L,{docs:h("services.mesh-multi-zone-service.href.docs")},{default:s(()=>[n(A,null,{default:s(()=>[n(V,{src:D(B(N),"/meshes/:mesh/mesh-multi-zone-services",{mesh:o.params.mesh},{page:o.params.page,size:o.params.size})},{loadable:s(({data:a})=>[n(C,{type:"services",items:(a==null?void 0:a.items)??[void 0]},{default:s(()=>[n(S,{headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.namespace"),label:"Namespace",key:"namespace"},...x("use zones")?[{...c.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...c.get("headers.addresses"),label:"Addresses",key:"addresses"},{...c.get("headers.ports"),label:"Ports",key:"ports"},{...c.get("headers.labels"),label:"Match Labels",key:"labels"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,"is-selected-row":e=>e.name===o.params.service,onChange:o.update,onResize:c.set},{name:s(({row:e})=>[n(T,{text:e.name},{default:s(()=>[n(d,{"data-action":"",to:{name:"mesh-multi-zone-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:o.params.page,size:o.params.size}}},{default:s(()=>[m(i(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[m(i(e.namespace),1)]),zone:s(({row:e})=>[e.labels&&e.labels["kuma.io/origin"]==="zone"&&e.labels["kuma.io/zone"]?(t(),u(_,{key:0},[e.labels["kuma.io/zone"]?(t(),p(d,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.labels["kuma.io/zone"]}}},{default:s(()=>[m(i(e.labels["kuma.io/zone"]),1)]),_:2},1032,["to"])):w("",!0)],64)):(t(),u(_,{key:1},[m(i(h("common.detail.none")),1)],64))]),addresses:s(({row:e})=>[n(z,null,{default:s(()=>[(t(!0),u(_,null,f(e.status.addresses,l=>(t(),u("span",{key:l.hostname},i(l.hostname),1))),128))]),_:2},1024)]),ports:s(({row:e})=>[n(z,null,{default:s(()=>[(t(!0),u(_,null,f(e.spec.ports,l=>(t(),p(g,{key:l.port,appearance:"info"},{default:s(()=>[m(i(l.port)+":"+i(l.targetPort)+"/"+i(l.appProtocol),1)]),_:2},1024))),128))]),_:2},1024)]),labels:s(({row:e})=>[n(z,null,{default:s(()=>[(t(!0),u(_,null,f(e.spec.selector.meshService.matchLabels,(l,v)=>(t(),p(g,{key:`${v}:${l}`,appearance:"info"},{default:s(()=>[m(i(v)+":"+i(l),1)]),_:2},1024))),128))]),_:2},1024)]),actions:s(({row:e})=>[n(k,null,{default:s(()=>[n(d,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[m(i(h("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"]),m(),a!=null&&a.items&&o.params.service?(t(),p(y,{key:0},{default:s(e=>[n(X,{onClose:l=>o.replace({name:"mesh-multi-zone-service-list-view",params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size}})},{default:s(()=>[(t(),p(M(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):w("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{Z as default};
