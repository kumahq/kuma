import{d as K,e as a,o as r,m as l,w as e,a as n,b as c,l as L,ap as S,A as X,$ as B,t as p,c as d,H as f,J as g,E as T,p as $}from"./index-BGYhp_E8.js";import{S as N}from"./SummaryView-DVexZK_i.js";const F=K({__name:"MeshMultiZoneServiceListView",setup(P){return(q,E)=>{const w=a("RouteTitle"),u=a("XAction"),z=a("KumaPort"),h=a("KTruncate"),C=a("XBadge"),b=a("XActionGroup"),y=a("RouterView"),V=a("DataCollection"),k=a("DataLoader"),A=a("KCard"),R=a("AppView"),x=a("RouteView");return r(),l(x,{name:"mesh-multi-zone-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:e(({route:t,t:_,uri:D,me:m})=>[n(w,{render:!1,title:_("services.routes.mesh-multi-zone-service-list-view.title")},null,8,["title"]),c(),n(R,{docs:_("services.mesh-multi-zone-service.href.docs")},{default:e(()=>[n(A,null,{default:e(()=>[n(k,{src:D(L(S),"/meshes/:mesh/mesh-multi-zone-services",{mesh:t.params.mesh},{page:t.params.page,size:t.params.size})},{loadable:e(({data:s})=>[n(V,{type:"services",items:(s==null?void 0:s.items)??[void 0]},{default:e(()=>[n(X,{headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.ports"),label:"Ports",key:"ports"},{...m.get("headers.labels"),label:"Selector",key:"labels"},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":t.params.page,"page-size":t.params.size,total:s==null?void 0:s.total,items:s==null?void 0:s.items,"is-selected-row":o=>o.name===t.params.service,onChange:t.update,onResize:m.set},{name:e(({row:o})=>[n(B,{text:o.name},{default:e(()=>[n(u,{"data-action":"",to:{name:"mesh-multi-zone-service-summary-view",params:{mesh:o.mesh,service:o.id},query:{page:t.params.page,size:t.params.size}}},{default:e(()=>[c(p(o.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),ports:e(({row:o})=>[n(h,null,{default:e(()=>[(r(!0),d(f,null,g(o.spec.ports,i=>(r(),l(z,{key:i.port,port:{...i,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),labels:e(({row:o})=>[n(h,null,{default:e(()=>[(r(!0),d(f,null,g(o.spec.selector.meshService.matchLabels,(i,v)=>(r(),l(C,{key:`${v}:${i}`,appearance:"info"},{default:e(()=>[c(p(v)+":"+p(i),1)]),_:2},1024))),128))]),_:2},1024)]),actions:e(({row:o})=>[n(b,null,{default:e(()=>[n(u,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:o.mesh,service:o.id}}},{default:e(()=>[c(p(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","is-selected-row","onChange","onResize"]),c(),s!=null&&s.items&&t.params.service?(r(),l(y,{key:0},{default:e(o=>[n(N,{onClose:i=>t.replace({name:"mesh-multi-zone-service-list-view",params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:e(()=>[(r(),l(T(o.Component),{items:s==null?void 0:s.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):$("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{F as default};
