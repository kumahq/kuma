import{d as V,a as l,o as a,b as i,w as e,e as r,a9 as x,f as t,t as n,m as p,c as m,F as u,G as v,a5 as c,q as b,a1 as _,_ as A}from"./index-DPw5bDvs.js";import{S as C}from"./StatusBadge-CgwF8Xry.js";const S={class:"stack-with-borders"},R=V({__name:"ZoneIngressSummaryView",props:{items:{}},setup(y){const g=y;return(B,D)=>{const z=l("RouteTitle"),f=l("RouterLink"),k=l("AppView"),h=l("DataCollection"),I=l("RouteView");return a(),i(I,{name:"zone-ingress-summary-view",params:{zoneIngress:""}},{default:e(({route:w,t:o})=>[r(h,{items:g.items,predicate:d=>d.id===w.params.zoneIngress,find:!0},{empty:e(()=>[r(x,null,{title:e(()=>[t(n(o("common.collection.summary.empty_title",{type:"ZoneIngress"})),1)]),default:e(()=>[t(),p("p",null,n(o("common.collection.summary.empty_message",{type:"ZoneIngress"})),1)]),_:2},1024)]),default:e(({items:d})=>[(a(!0),m(u,null,v([d[0]],s=>(a(),i(k,{key:s.id},{title:e(()=>[p("h2",null,[r(f,{to:{name:"zone-ingress-detail-view",params:{zone:s.zoneIngress.zone,zoneIngress:s.id}}},{default:e(()=>[r(z,{title:o("zone-ingresses.routes.item.title",{name:s.name})},null,8,["title"])]),_:2},1032,["to"])])]),default:e(()=>[t(),p("div",S,[r(c,{layout:"horizontal"},{title:e(()=>[t(n(o("http.api.property.status")),1)]),body:e(()=>[r(C,{status:s.state},null,8,["status"])]),_:2},1024),t(),s.namespace.length>0?(a(),i(c,{key:0,layout:"horizontal"},{title:e(()=>[t(n(o("data-planes.routes.item.namespace")),1)]),body:e(()=>[t(n(s.namespace),1)]),_:2},1024)):b("",!0),t(),r(c,{layout:"horizontal"},{title:e(()=>[t(n(o("http.api.property.address")),1)]),body:e(()=>[s.zoneIngress.socketAddress.length>0?(a(),i(_,{key:0,text:s.zoneIngress.socketAddress},null,8,["text"])):(a(),m(u,{key:1},[t(n(o("common.detail.none")),1)],64))]),_:2},1024),t(),r(c,{layout:"horizontal"},{title:e(()=>[t(n(o("http.api.property.advertisedAddress")),1)]),body:e(()=>[s.zoneIngress.advertisedSocketAddress.length>0?(a(),i(_,{key:0,text:s.zoneIngress.advertisedSocketAddress},null,8,["text"])):(a(),m(u,{key:1},[t(n(o("common.detail.none")),1)],64))]),_:2},1024)])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),L=A(R,[["__scopeId","data-v-b43ed833"]]);export{L as default};
