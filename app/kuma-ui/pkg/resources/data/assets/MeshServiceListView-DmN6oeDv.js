import{d as K,r as c,o,m as p,w as s,b as n,e as i,l as L,ax as B,A as N,a2 as y,t as r,c as _,L as u,M as f,E as S,p as X}from"./index-CeTpyiyE.js";import{S as $}from"./SummaryView-wGXCefwD.js";const F=K({__name:"MeshServiceListView",setup(P){return(q,E)=>{const z=c("RouteTitle"),h=c("XAction"),g=c("KTruncate"),v=c("KBadge"),k=c("XActionGroup"),C=c("RouterView"),V=c("DataCollection"),b=c("DataLoader"),A=c("KCard"),x=c("AppView"),R=c("RouteView");return o(),p(R,{name:"mesh-service-list-view",params:{page:1,size:50,mesh:"",service:""}},{default:s(({route:t,t:d,can:T,uri:D,me:m})=>[n(z,{render:!1,title:d("services.routes.mesh-service-list-view.title")},null,8,["title"]),i(),n(x,{docs:d("services.mesh-service.href.docs")},{default:s(()=>[n(A,null,{default:s(()=>[n(b,{src:D(L(B),"/meshes/:mesh/mesh-services",{mesh:t.params.mesh},{page:t.params.page,size:t.params.size})},{loadable:s(({data:a})=>[n(V,{type:"services",items:(a==null?void 0:a.items)??[void 0],page:t.params.page,"page-size":t.params.size,total:a==null?void 0:a.total,onChange:t.update},{default:s(()=>[n(N,{"data-testid":"service-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.namespace"),label:"Namespace",key:"namespace"},...T("use zones")?[{...m.get("headers.zone"),label:"Zone",key:"zone"}]:[],{...m.get("headers.addresses"),label:"Addresses",key:"addresses"},{...m.get("headers.ports"),label:"Ports",key:"ports"},{...m.get("headers.tags"),label:"Tags",key:"tags"},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,"is-selected-row":e=>e.name===t.params.service,onResize:m.set},{name:s(({row:e})=>[n(y,{text:e.name},{default:s(()=>[n(h,{"data-action":"",to:{name:"mesh-service-summary-view",params:{mesh:e.mesh,service:e.id},query:{page:t.params.page,size:t.params.size}}},{default:s(()=>[i(r(e.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),namespace:s(({row:e})=>[i(r(e.namespace),1)]),zone:s(({row:e})=>[e.zone?(o(),p(h,{key:0,to:{name:"zone-cp-detail-view",params:{zone:e.zone}}},{default:s(()=>[i(r(e.zone),1)]),_:2},1032,["to"])):(o(),_(u,{key:1},[i(r(d("common.detail.none")),1)],64))]),addresses:s(({row:e})=>[e.status.addresses.length===1?(o(),p(y,{key:0,text:e.status.addresses[0].hostname},{default:s(()=>[i(r(e.status.addresses[0].hostname),1)]),_:2},1032,["text"])):(o(),p(g,{key:1},{default:s(()=>[(o(!0),_(u,null,f(e.status.addresses,l=>(o(),_("span",{key:l.hostname},r(l.hostname),1))),128))]),_:2},1024))]),ports:s(({row:e})=>[n(g,null,{default:s(()=>[(o(!0),_(u,null,f(e.spec.ports,l=>(o(),p(v,{key:l.port,appearance:"info"},{default:s(()=>[i(r(l.port)+":"+r(l.targetPort)+"/"+r(l.appProtocol),1)]),_:2},1024))),128))]),_:2},1024)]),tags:s(({row:e})=>[n(g,null,{default:s(()=>[(o(!0),_(u,null,f(e.spec.selector.dataplaneTags,(l,w)=>(o(),p(v,{key:`${w}:${l}`,appearance:"info"},{default:s(()=>[i(r(w)+":"+r(l),1)]),_:2},1024))),128))]),_:2},1024)]),actions:s(({row:e})=>[n(k,null,{default:s(()=>[n(h,{to:{name:"mesh-service-detail-view",params:{mesh:e.mesh,service:e.id}}},{default:s(()=>[i(r(d("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),i(),a!=null&&a.items&&t.params.service?(o(),p(C,{key:0},{default:s(e=>[n($,{onClose:l=>t.replace({name:"mesh-service-list-view",params:{mesh:t.params.mesh},query:{page:t.params.page,size:t.params.size}})},{default:s(()=>[(o(),p(S(e.Component),{items:a==null?void 0:a.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):X("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{F as default};
