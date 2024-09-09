import{d as v,r as s,o as c,m as d,w as e,p as m,e as l,b as a,T as E,k as f,l as R,aA as b,A as L,t as u,a2 as S,c as x,L as D,S as T,E as B}from"./index-sgqUZBhH.js";import{S as N}from"./SummaryView-lo6vVtOH.js";const X=["innerHTML"],F=v({__name:"ZoneEgressListView",props:{data:{}},setup(M){return(q,G)=>{const p=s("RouteTitle"),_=s("XAction"),w=s("XActionGroup"),k=s("RouterView"),y=s("DataCollection"),g=s("DataLoader"),h=s("KCard"),A=s("AppView"),C=s("RouteView");return c(),d(C,{name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:e(({route:n,t:i,me:r,uri:V,can:z})=>[z("use zones")?(c(),d(p,{key:0,render:!1,title:i("zone-egresses.routes.items.title")},null,8,["title"])):m("",!0),l(),a(A,{docs:i("zone-egresses.href.docs")},E({default:e(()=>[l(),f("div",{innerHTML:i("zone-egresses.routes.items.intro",{},{defaultMessage:""})},null,8,X),l(),a(h,null,{default:e(()=>[a(g,{src:V(R(b),"/zone-cps/:name/egresses",{name:n.params.zone||"*"},{page:1,size:100})},{loadable:e(({data:o})=>[a(y,{type:"zone-egresses",items:(o==null?void 0:o.items)??[void 0],total:o==null?void 0:o.total,onChange:n.update},{default:e(()=>[a(L,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{...r.get("headers.name"),label:"Name",key:"name"},{...r.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...r.get("headers.status"),label:"Status",key:"status"},{...r.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,"is-selected-row":t=>t.name===n.params.zoneEgress,onResize:r.set},{name:e(({row:t})=>[a(_,{"data-action":"",to:{name:"zone-egress-summary-view",params:{zone:n.params.zone,zoneEgress:t.id},query:{page:1,size:100}}},{default:e(()=>[l(u(t.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:t})=>[t.zoneEgress.socketAddress.length>0?(c(),d(S,{key:0,text:t.zoneEgress.socketAddress},null,8,["text"])):(c(),x(D,{key:1},[l(u(i("common.collection.none")),1)],64))]),status:e(({row:t})=>[a(T,{status:t.state},null,8,["status"])]),actions:e(({row:t})=>[a(w,null,{default:e(()=>[a(_,{to:{name:"zone-egress-detail-view",params:{zoneEgress:t.id}}},{default:e(()=>[l(u(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),l(),a(k,null,{default:e(({Component:t})=>[n.child()?(c(),d(N,{key:0,onClose:H=>n.replace({name:"zone-egress-list-view",params:{zone:n.params.zone},query:{page:1,size:100}})},{default:e(()=>[typeof o<"u"?(c(),d(B(t),{key:0,items:o.items},null,8,["items"])):m("",!0)]),_:2},1032,["onClose"])):m("",!0)]),_:2},1024)]),_:2},1032,["items","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},[z("use zones")?void 0:{name:"title",fn:e(()=>[f("h1",null,[a(p,{title:i("zone-egresses.routes.items.title")},null,8,["title"])])]),key:"0"}]),1032,["docs"])]),_:1})}}});export{F as default};
