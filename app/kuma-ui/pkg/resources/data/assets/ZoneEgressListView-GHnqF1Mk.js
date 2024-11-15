import{d as h,e as n,o as d,m,w as e,p as u,b as i,a as s,Q as R,k as w,l as b,ar as S,A as x,t as p,$ as D,c as L,H as B,S as N,E as T}from"./index-CjjKwNo4.js";import{S as X}from"./SummaryView-BhSKxYbl.js";const H=["innerHTML"],F=h({__name:"ZoneEgressListView",props:{data:{}},setup(M){return($,a)=>{const _=n("RouteTitle"),z=n("XAction"),k=n("XActionGroup"),y=n("RouterView"),g=n("DataCollection"),A=n("DataLoader"),C=n("KCard"),V=n("AppView"),v=n("RouteView");return d(),m(v,{name:"zone-egress-list-view",params:{zone:"",zoneEgress:""}},{default:e(({route:l,t:r,me:c,uri:E,can:f})=>[f("use zones")?(d(),m(_,{key:0,render:!1,title:r("zone-egresses.routes.items.title")},null,8,["title"])):u("",!0),a[6]||(a[6]=i()),s(V,{docs:r("zone-egresses.href.docs")},R({default:e(()=>[a[4]||(a[4]=i()),w("div",{innerHTML:r("zone-egresses.routes.items.intro",{},{defaultMessage:""})},null,8,H),a[5]||(a[5]=i()),s(C,null,{default:e(()=>[s(A,{src:E(b(S),"/zone-cps/:name/egresses",{name:l.params.zone||"*"},{page:1,size:100})},{loadable:e(({data:o})=>[s(g,{type:"zone-egresses",items:(o==null?void 0:o.items)??[void 0],total:o==null?void 0:o.total,onChange:l.update},{default:e(()=>[s(x,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...c.get("headers.status"),label:"Status",key:"status"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:o==null?void 0:o.items,"is-selected-row":t=>t.name===l.params.zoneEgress,onResize:c.set},{name:e(({row:t})=>[s(z,{"data-action":"",to:{name:"zone-egress-summary-view",params:{zone:l.params.zone,zoneEgress:t.id},query:{page:1,size:100}}},{default:e(()=>[i(p(t.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:t})=>[t.zoneEgress.socketAddress.length>0?(d(),m(D,{key:0,text:t.zoneEgress.socketAddress},null,8,["text"])):(d(),L(B,{key:1},[i(p(r("common.collection.none")),1)],64))]),status:e(({row:t})=>[s(N,{status:t.state},null,8,["status"])]),actions:e(({row:t})=>[s(k,null,{default:e(()=>[s(z,{to:{name:"zone-egress-detail-view",params:{zoneEgress:t.id}}},{default:e(()=>[i(p(r("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),a[3]||(a[3]=i()),s(y,null,{default:e(({Component:t})=>[l.child()?(d(),m(X,{key:0,onClose:q=>l.replace({name:"zone-egress-list-view",params:{zone:l.params.zone},query:{page:1,size:100}})},{default:e(()=>[typeof o<"u"?(d(),m(T(t),{key:0,items:o.items},null,8,["items"])):u("",!0)]),_:2},1032,["onClose"])):u("",!0)]),_:2},1024)]),_:2},1032,["items","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},[f("use zones")?void 0:{name:"title",fn:e(()=>[w("h1",null,[s(_,{title:r("zone-egresses.routes.items.title")},null,8,["title"])])]),key:"0"}]),1032,["docs"])]),_:1})}}});export{F as default};
