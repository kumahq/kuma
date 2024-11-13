import{d as K,e as n,o as l,k as m,w as s,a as o,b as r,j as X,aq as B,A as L,$ as N,t as c,c as v,F as g,G as S,C as T,l as q}from"./index-loxRIpgb.js";import{S as G}from"./SummaryView-Bl9n50Fq.js";const Z=K({__name:"MeshServiceListView",setup($){return(F,j)=>{const w=n("RouteTitle"),u=n("XAction"),f=n("XBadge"),z=n("KumaPort"),y=n("KTruncate"),C=n("XActionGroup"),k=n("RouterView"),b=n("DataCollection"),V=n("DataLoader"),x=n("KCard"),A=n("AppView"),P=n("RouteView");return l(),m(P,{name:"mesh-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:s(({route:t,t:_,can:R,uri:D,me:i})=>[o(w,{render:!1,title:_("services.routes.mesh-service-list-view.title")},null,8,["title"]),r(),o(A,{docs:_("services.mesh-service.href.docs")},{default:s(()=>[o(x,null,{default:s(()=>[o(V,{src:D(X(B),"/meshes/:mesh/mesh-services",{mesh:t.params.mesh},{page:t.params.page,size:t.params.size})},{loadable:s(({data:a})=>[o(b,{type:"services",items:(a==null?void 0:a.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:a==null?void 0:a.total,onChange:t.update},{default:s(()=>[o(L,{"data-testid":"service-collection",headers:[{...i.get("headers.name"),label:"Name",key:"name"},{...i.get("headers.namespace"),label:"Namespace",key:"namespace"},...R("use zones")?[{...i.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...i.get("headers.state"),label:"State",key:"state"},{...i.get("headers.status"),label:"DP proxies (connected / healthy / total)",key:"status"},{...i.get("headers.ports"),label:"Ports",key:"ports"},{...i.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,"is-selected-row":e=>e.name===t.params.service,onResize:i.set},{name:s(({row:e})=>[o(N,{text:e.name},{default:s(()=>[o(u,{"data-action":"",to:{name:"mesh-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:t.params.page,size:t.params.size}}},{default:s(()=>[r(c(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[r(c(e.namespace),1)]),zone:s(({row:e})=>[e.zone?(l(),m(u,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[r(c(e.zone),1)]),_:2},1032,["to"])):(l(),v(g,{key:1},[r(c(_("common.detail.none")),1)],64))]),state:s(({row:e})=>[o(f,{appearance:e.spec.state==="Available"?"success":"danger"},{default:s(()=>[r(c(e.spec.state),1)]),_:2},1032,["appearance"])]),status:s(({row:e})=>{var p,h,d;return[r(c((p=e.status.dataplaneProxies)==null?void 0:p.connected)+" / "+c((h=e.status.dataplaneProxies)==null?void 0:h.healthy)+" / "+c((d=e.status.dataplaneProxies)==null?void 0:d.total),1)]}),ports:s(({row:e})=>[o(y,null,{default:s(()=>[(l(!0),v(g,null,S(e.spec.ports,p=>(l(),m(z,{key:p.port,port:{...p,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),actions:s(({row:e})=>[o(C,null,{default:s(()=>[o(u,{to:{name:"mesh-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[r(c(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),r(),a!=null&&a.items&&t.params.service?(l(),m(k,{key:0},{default:s(e=>[o(G,{onClose:p=>t.replace({name:"mesh-service-list-view",params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:s(()=>[(l(),m(T(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):q("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{Z as default};
