import{d as K,r as c,o as i,m,w as s,b as o,e as l,l as L,ax as P,A as S,a2 as B,t as n,c as g,L as w,M as N,E as T,p as X}from"./index-BZS2OlAU.js";import{S as $}from"./SummaryView-WylnThA7.js";const Z=K({__name:"MeshServiceListView",setup(q){return(E,G)=>{const f=c("RouteTitle"),u=c("XAction"),h=c("KBadge"),z=c("KTruncate"),y=c("XActionGroup"),C=c("RouterView"),k=c("DataCollection"),b=c("DataLoader"),x=c("KCard"),V=c("AppView"),A=c("RouteView");return i(),m(A,{name:"mesh-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:s(({route:t,t:_,can:R,uri:D,me:p})=>[o(f,{render:!1,title:_("services.routes.mesh-service-list-view.title")},null,8,["title"]),l(),o(V,{docs:_("services.mesh-service.href.docs")},{default:s(()=>[o(x,null,{default:s(()=>[o(b,{src:D(L(P),"/meshes/:mesh/mesh-services",{mesh:t.params.mesh},{page:t.params.page,size:t.params.size})},{loadable:s(({data:a})=>[o(k,{type:"services",items:(a==null?void 0:a.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:a==null?void 0:a.total,onChange:t.update},{default:s(()=>[o(S,{"data-testid":"service-collection",headers:[{...p.get("headers.name"),label:"Name",key:"name"},{...p.get("headers.namespace"),label:"Namespace",key:"namespace"},...R("use zones")?[{...p.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...p.get("headers.state"),label:"State",key:"state"},{...p.get("headers.status"),label:"DP proxies (connected / healthy / total)",key:"status"},{...p.get("headers.ports"),label:"Ports",key:"ports"},{...p.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,"is-selected-row":e=>e.name===t.params.service,onResize:p.set},{name:s(({row:e})=>[o(B,{text:e.name},{default:s(()=>[o(u,{"data-action":"",to:{name:"mesh-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:t.params.page,size:t.params.size}}},{default:s(()=>[l(n(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[l(n(e.namespace),1)]),zone:s(({row:e})=>[e.zone?(i(),m(u,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[l(n(e.zone),1)]),_:2},1032,["to"])):(i(),g(w,{key:1},[l(n(_("common.detail.none")),1)],64))]),state:s(({row:e})=>[o(h,{appearance:e.spec.state==="Available"?"success":"danger"},{default:s(()=>[l(n(e.spec.state),1)]),_:2},1032,["appearance"])]),status:s(({row:e})=>{var r,d,v;return[l(n((r=e.status.dataplaneProxies)==null?void 0:r.connected)+" / "+n((d=e.status.dataplaneProxies)==null?void 0:d.healthy)+" / "+n((v=e.status.dataplaneProxies)==null?void 0:v.total),1)]}),ports:s(({row:e})=>[o(z,null,{default:s(()=>[(i(!0),g(w,null,N(e.spec.ports,r=>(i(),m(h,{key:r.port,appearance:"info"},{default:s(()=>[l(n(r.port)+"/"+n(r.appProtocol)+n(r.name&&r.name!==String(r.port)?` (${r.name})`:""),1)]),_:2},1024))),128))]),_:2},1024)]),actions:s(({row:e})=>[o(y,null,{default:s(()=>[o(u,{to:{name:"mesh-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[l(n(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l(),a!=null&&a.items&&t.params.service?(i(),m(C,{key:0},{default:s(e=>[o($,{onClose:r=>t.replace({name:"mesh-service-list-view",params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:s(()=>[(i(),m(T(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):X("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{Z as default};
