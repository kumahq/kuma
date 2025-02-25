import{d as V,r as n,o as a,m as p,w as t,b as o,e as r,p as X,au as b,C as R,t as m,c as _,F as z,S as I,K as B,q as D}from"./index-Cm0u77zo.js";import{S as N}from"./SummaryView-2SryJn68.js";const E=V({__name:"ZoneIngressListView",setup(T){return(L,i)=>{const y=n("RouteTitle"),k=n("XI18n"),g=n("XAction"),u=n("XCopyButton"),w=n("XActionGroup"),A=n("RouterView"),f=n("DataCollection"),C=n("DataLoader"),v=n("XCard"),x=n("AppView"),h=n("RouteView");return a(),p(h,{name:"zone-ingress-list-view",params:{page:1,size:Number,zone:"",proxy:""}},{default:t(({route:s,t:d,me:l,uri:S})=>[o(y,{render:!1,title:d("zone-ingresses.routes.items.title")},null,8,["title"]),i[6]||(i[6]=r()),o(x,{docs:d("zone-ingresses.href.docs")},{default:t(()=>[o(k,{path:"zone-ingresses.routes.items.intro","default-path":"common.i18n.ignore-error"}),i[5]||(i[5]=r()),o(v,null,{default:t(()=>[o(C,{src:S(X(b),"/zone-cps/:name/ingresses",{name:s.params.zone},{page:s.params.page,size:s.params.size}),variant:"list"},{default:t(({data:c})=>[o(f,{type:"zone-ingresses",items:c.items,page:s.params.page,"page-size":s.params.size,total:c.total,onChange:s.update},{default:t(()=>[o(R,{class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{...l.get("headers.name"),label:"Name",key:"name"},{...l.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...l.get("headers.advertisedSocketAddress"),label:"Advertised address",key:"advertisedSocketAddress"},{...l.get("headers.status"),label:"Status",key:"status"},{...l.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:c.items,"is-selected-row":e=>e.name===s.params.proxy,onResize:l.set},{name:t(({row:e})=>[o(g,{"data-action":"",to:{name:"zone-ingress-summary-view",params:{zone:s.params.zone,proxy:e.id,proxyType:"ingresses"},query:{page:s.params.page,size:s.params.size}}},{default:t(()=>[r(m(e.name),1)]),_:2},1032,["to"])]),socketAddress:t(({row:e})=>[e.zoneIngress.socketAddress.length>0?(a(),p(u,{key:0,text:e.zoneIngress.socketAddress},null,8,["text"])):(a(),_(z,{key:1},[r(m(d("common.collection.none")),1)],64))]),advertisedSocketAddress:t(({row:e})=>[e.zoneIngress.advertisedSocketAddress.length>0?(a(),p(u,{key:0,text:e.zoneIngress.advertisedSocketAddress},null,8,["text"])):(a(),_(z,{key:1},[r(m(d("common.collection.none")),1)],64))]),status:t(({row:e})=>[o(I,{status:e.state},null,8,["status"])]),actions:t(({row:e})=>[o(w,null,{default:t(()=>[o(g,{to:{name:"zone-ingress-detail-view",params:{proxy:e.id,proxyType:"ingresses"}}},{default:t(()=>[r(m(d("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),i[4]||(i[4]=r()),s.child()?(a(),p(A,{key:0},{default:t(({Component:e})=>[o(N,{onClose:q=>s.replace({name:"zone-ingress-list-view",params:{zone:s.params.zone},query:{page:s.params.page,size:s.params.size}})},{default:t(()=>[(a(),p(B(e),{items:c.items},null,8,["items"]))]),_:2},1032,["onClose"])]),_:2},1024)):D("",!0)]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{E as default};
