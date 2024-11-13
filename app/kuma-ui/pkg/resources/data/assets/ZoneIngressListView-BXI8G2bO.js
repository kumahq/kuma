import{d as S,e as a,o as d,k as i,w as t,a as s,b as r,i as I,j as b,at as x,A as R,t as m,$ as u,c as _,F as z,S as D,C as L,l as k}from"./index-loxRIpgb.js";import{S as B}from"./SummaryView-Bl9n50Fq.js";const N=["innerHTML"],G=S({__name:"ZoneIngressListView",props:{data:{}},setup(T){return(X,M)=>{const w=a("RouteTitle"),p=a("XAction"),f=a("XActionGroup"),A=a("RouterView"),g=a("DataCollection"),y=a("DataLoader"),h=a("KCard"),v=a("AppView"),C=a("RouteView");return d(),i(C,{name:"zone-ingress-list-view",params:{zone:"",zoneIngress:""}},{default:t(({route:n,t:l,me:c,uri:V})=>[s(w,{render:!1,title:l("zone-ingresses.routes.items.title")},null,8,["title"]),r(),s(v,{docs:l("zone-ingresses.href.docs")},{default:t(()=>[I("div",{innerHTML:l("zone-ingresses.routes.items.intro",{},{defaultMessage:""})},null,8,N),r(),s(h,null,{default:t(()=>[s(y,{src:V(b(x),"/zone-cps/:name/ingresses",{name:n.params.zone},{page:1,size:100})},{loadable:t(({data:o})=>[s(g,{type:"zone-ingresses",items:(o==null?void 0:o.items)??[void 0],total:o==null?void 0:o.total,onChange:n.update},{default:t(()=>[s(R,{class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...c.get("headers.advertisedSocketAddress"),label:"Advertised address",key:"advertisedSocketAddress"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,"is-selected-row":e=>e.name===n.params.zoneIngress,onResize:c.set},{name:t(({row:e})=>[s(p,{"data-action":"",to:{name:"zone-ingress-summary-view",params:{zone:n.params.zone,zoneIngress:e.id},query:{page:1,size:100}}},{default:t(()=>[r(m(e.name),1)]),_:2},1032,["to"])]),socketAddress:t(({row:e})=>[e.zoneIngress.socketAddress.length>0?(d(),i(u,{key:0,text:e.zoneIngress.socketAddress},null,8,["text"])):(d(),_(z,{key:1},[r(m(l("common.collection.none")),1)],64))]),advertisedSocketAddress:t(({row:e})=>[e.zoneIngress.advertisedSocketAddress.length>0?(d(),i(u,{key:0,text:e.zoneIngress.advertisedSocketAddress},null,8,["text"])):(d(),_(z,{key:1},[r(m(l("common.collection.none")),1)],64))]),status:t(({row:e})=>[s(D,{status:e.state},null,8,["status"])]),actions:t(({row:e})=>[s(f,null,{default:t(()=>[s(p,{to:{name:"zone-ingress-detail-view",params:{zoneIngress:e.id}}},{default:t(()=>[r(m(l("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),r(),n.child()?(d(),i(A,{key:0},{default:t(({Component:e})=>[s(B,{onClose:$=>n.replace({name:"zone-ingress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:t(()=>[typeof o<"u"?(d(),i(L(e),{key:0,items:o.items},null,8,["items"])):k("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):k("",!0)]),_:2},1032,["items","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{G as default};
