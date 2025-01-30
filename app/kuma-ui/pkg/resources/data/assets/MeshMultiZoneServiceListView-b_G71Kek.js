import{d as D,r as n,o as r,q as l,w as e,b as a,e as p,p as N,aq as S,B as q,t as u,c as g,M as f,N as w,I as P,s as M}from"./index-Du84oSnm.js";import{S as T}from"./SummaryView-Cd8oe3uM.js";const F=D({__name:"MeshMultiZoneServiceListView",setup($){return(G,c)=>{const y=n("RouteTitle"),h=n("XAction"),z=n("XCopyButton"),C=n("KumaPort"),v=n("XLayout"),b=n("XBadge"),X=n("XActionGroup"),V=n("RouterView"),k=n("DataCollection"),A=n("DataLoader"),B=n("XCard"),L=n("AppView"),R=n("RouteView");return r(),l(R,{name:"mesh-multi-zone-service-list-view",params:{page:1,size:Number,mesh:"",service:""}},{default:e(({route:o,t:_,uri:x,me:m})=>[a(y,{render:!1,title:_("services.routes.mesh-multi-zone-service-list-view.title")},null,8,["title"]),c[4]||(c[4]=p()),a(L,{docs:_("services.mesh-multi-zone-service.href.docs")},{default:e(()=>[a(B,null,{default:e(()=>[a(A,{src:x(N(S),"/meshes/:mesh/mesh-multi-zone-services",{mesh:o.params.mesh},{page:o.params.page,size:o.params.size})},{loadable:e(({data:s})=>[a(k,{type:"services",items:(s==null?void 0:s.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:s==null?void 0:s.total,onChange:o.update},{default:e(()=>[a(q,{headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.ports"),label:"Ports",key:"ports"},{...m.get("headers.labels"),label:"Selector",key:"labels"},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:s==null?void 0:s.items,"is-selected-row":t=>t.name===o.params.service,onResize:m.set},{name:e(({row:t})=>[a(z,{text:t.name},{default:e(()=>[a(h,{"data-action":"",to:{name:"mesh-multi-zone-service-summary-view",params:{mesh:t.mesh,service:t.id},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[p(u(t.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),ports:e(({row:t})=>[a(v,{type:"separated",truncate:""},{default:e(()=>[(r(!0),g(f,null,w(t.spec.ports,i=>(r(),l(C,{key:i.port,port:{...i,targetPort:void 0}},null,8,["port"]))),128))]),_:2},1024)]),labels:e(({row:t})=>[a(v,{type:"separated",truncate:""},{default:e(()=>[(r(!0),g(f,null,w(t.spec.selector.meshService.matchLabels,(i,d)=>(r(),l(b,{key:`${d}:${i}`,appearance:"info"},{default:e(()=>[p(u(d)+":"+u(i),1)]),_:2},1024))),128))]),_:2},1024)]),actions:e(({row:t})=>[a(X,null,{default:e(()=>[a(h,{to:{name:"mesh-multi-zone-service-detail-view",params:{mesh:t.mesh,service:t.id}}},{default:e(()=>[p(u(_("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),c[3]||(c[3]=p()),s!=null&&s.items&&o.params.service?(r(),l(V,{key:0},{default:e(t=>[a(T,{onClose:i=>o.replace({name:"mesh-multi-zone-service-list-view",params:{mesh:o.params.mesh},query:{page:o.params.page,size:o.params.size}})},{default:e(()=>[(r(),l(P(t.Component),{items:s==null?void 0:s.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):M("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{F as default};
