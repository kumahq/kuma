import{d as h,r as a,o as i,q as c,w as e,s as d,e as r,b as o,T as B,p as R,as as T,B as b,t as u,c as S,M as D,S as N,I as E,m as L}from"./index-DxWkv34s.js";import{S as q}from"./SummaryView-DD4ymD2U.js";const $=h({__name:"ZoneEgressListView",props:{data:{}},setup(I){return(G,s)=>{const _=a("RouteTitle"),f=a("XI18n"),y=a("XAction"),w=a("XCopyButton"),k=a("XActionGroup"),C=a("RouterView"),g=a("DataCollection"),x=a("DataLoader"),A=a("XCard"),V=a("AppView"),v=a("RouteView");return i(),c(v,{name:"zone-egress-list-view",params:{zone:"",proxy:"",proxyType:""}},{default:e(({route:l,t:p,me:m,uri:X,can:z})=>[z("use zones")?(i(),c(_,{key:0,render:!1,title:p("zone-egresses.routes.items.title")},null,8,["title"])):d("",!0),s[6]||(s[6]=r()),o(V,{docs:p("zone-egresses.href.docs")},B({default:e(()=>[s[4]||(s[4]=r()),o(f,{path:"zone-egresses.routes.items.intro","default-path":"common.i18n.ignore-error"}),s[5]||(s[5]=r()),o(A,null,{default:e(()=>[o(x,{src:X(R(T),"/zone-cps/:name/egresses",{name:l.params.zone||"*"},{page:1,size:100})},{loadable:e(({data:n})=>[o(g,{type:"zone-egresses",items:(n==null?void 0:n.items)??[void 0],total:n==null?void 0:n.total,onChange:l.update},{default:e(()=>[o(b,{class:"zone-egress-collection","data-testid":"zone-egress-collection",headers:[{...m.get("headers.name"),label:"Name",key:"name"},{...m.get("headers.socketAddress"),label:"Address",key:"socketAddress"},{...m.get("headers.status"),label:"Status",key:"status"},{...m.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:n==null?void 0:n.items,"is-selected-row":t=>t.name===l.params.proxy,onResize:m.set},{name:e(({row:t})=>[o(y,{"data-action":"",to:{name:"zone-egress-summary-view",params:{zone:l.params.zone,proxy:t.id},query:{page:1,size:100}}},{default:e(()=>[r(u(t.name),1)]),_:2},1032,["to"])]),socketAddress:e(({row:t})=>[t.zoneEgress.socketAddress.length>0?(i(),c(w,{key:0,text:t.zoneEgress.socketAddress},null,8,["text"])):(i(),S(D,{key:1},[r(u(p("common.collection.none")),1)],64))]),status:e(({row:t})=>[o(N,{status:t.state},null,8,["status"])]),actions:e(({row:t})=>[o(k,null,{default:e(()=>[o(y,{to:{name:"zone-egress-detail-view",params:{proxyType:"egresses",proxy:t.id}}},{default:e(()=>[r(u(p("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"]),s[3]||(s[3]=r()),o(C,null,{default:e(({Component:t})=>[l.child()?(i(),c(q,{key:0,onClose:F=>l.replace({name:"zone-egress-list-view",params:{zone:l.params.zone,proxyType:l.params.proxyType},query:{page:1,size:100}})},{default:e(()=>[typeof n<"u"?(i(),c(E(t),{key:0,items:n.items},null,8,["items"])):d("",!0)]),_:2},1032,["onClose"])):d("",!0)]),_:2},1024)]),_:2},1032,["items","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},[z("use zones")?void 0:{name:"title",fn:e(()=>[L("h1",null,[o(_,{title:p("zone-egresses.routes.items.title")},null,8,["title"])])]),key:"0"}]),1032,["docs"])]),_:1})}}});export{$ as default};
