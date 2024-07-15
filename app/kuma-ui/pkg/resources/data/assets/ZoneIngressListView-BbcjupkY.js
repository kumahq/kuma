import{d as S,r as t,o as r,m as d,w as s,b as n,e as i,k as b,l as I,at as x,A as R,t as m,T as u,c as _,F as g,S as D,E as L,p as z}from"./index-DxrN05KS.js";import{S as T}from"./SummaryView-jGgfzaaj.js";const B=["innerHTML"],G=S({__name:"ZoneIngressListView",setup(N){return(X,M)=>{const k=t("RouteTitle"),p=t("XAction"),w=t("XActionGroup"),f=t("RouterView"),A=t("DataCollection"),y=t("DataLoader"),h=t("KCard"),v=t("AppView"),C=t("RouteView");return r(),d(C,{name:"zone-ingress-list-view",params:{zone:"",zoneIngress:""}},{default:s(({route:a,t:l,me:c,uri:V})=>[n(k,{render:!1,title:l("zone-ingresses.routes.items.title")},null,8,["title"]),i(),n(v,{docs:l("zone-ingresses.href.docs")},{default:s(()=>[b("div",{innerHTML:l("zone-ingresses.routes.items.intro",{},{defaultMessage:""})},null,8,B),i(),n(h,null,{default:s(()=>[n(y,{src:V(I(x),"/zone-cps/:name/ingresses",{name:a.params.zone},{page:1,size:100})},{loadable:s(({data:o})=>[n(A,{type:"zone-ingresses",items:(o==null?void 0:o.items)??[void 0]},{default:s(()=>[n(R,{class:"zone-ingress-collection","data-testid":"zone-ingress-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...c.get("headers.advertisedSocketAddress"),label:"Advertised address",key:"advertisedSocketAddress"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":1,"page-size":100,total:o==null?void 0:o.total,items:o==null?void 0:o.items,"is-selected-row":e=>e.name===a.params.zoneIngress,onChange:a.update,onResize:c.set},{name:s(({row:e})=>[n(p,{"data-action":"",to:{name:"zone-ingress-summary-view",params:{zone:a.params.zone,zoneIngress:e.id},query:{page:1,size:100}}},{default:s(()=>[i(m(e.name),1)]),_:2},1032,["to"])]),socketAddress:s(({row:e})=>[e.zoneIngress.socketAddress.length>0?(r(),d(u,{key:0,text:e.zoneIngress.socketAddress},null,8,["text"])):(r(),_(g,{key:1},[i(m(l("common.collection.none")),1)],64))]),advertisedSocketAddress:s(({row:e})=>[e.zoneIngress.advertisedSocketAddress.length>0?(r(),d(u,{key:0,text:e.zoneIngress.advertisedSocketAddress},null,8,["text"])):(r(),_(g,{key:1},[i(m(l("common.collection.none")),1)],64))]),status:s(({row:e})=>[n(D,{status:e.state},null,8,["status"])]),actions:s(({row:e})=>[n(w,null,{default:s(()=>[n(p,{to:{name:"zone-ingress-detail-view",params:{zoneIngress:e.id}}},{default:s(()=>[i(m(l("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","total","items","is-selected-row","onChange","onResize"]),i(),a.child()?(r(),d(f,{key:0},{default:s(({Component:e})=>[n(T,{onClose:q=>a.replace({name:"zone-ingress-list-view",params:{zone:a.params.zone},query:{page:1,size:100}})},{default:s(()=>[typeof o<"u"?(r(),d(L(e),{key:0,items:o.items},null,8,["items"])):z("",!0)]),_:2},1032,["onClose"])]),_:2},1024)):z("",!0)]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{G as default};
