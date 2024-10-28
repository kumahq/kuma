import{d as f,e as _,o as n,m as u,w as t,a as l,k as p,Y as r,b as e,t as s,c,H as m,J as b,p as h,r as V}from"./index-C4IVBmnO.js";const g={class:"stack-with-borders"},C={class:"mt-8 stack-with-borders"},B=f({__name:"SubscriptionSummaryOverviewView",props:{data:{}},setup(v){const o=v;return(w,S)=>{const z=_("AppView"),I=_("RouteView");return n(),u(I,{name:"subscription-summary-overview-view"},{default:t(({t:a})=>[l(z,null,{default:t(()=>{var y;return[p("div",g,[l(r,{layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.version")),1)]),body:t(()=>{var i,d;return[(n(!0),c(m,null,b([(d=(i=o.data.version)==null?void 0:i.kumaCp)==null?void 0:d.version],k=>(n(),c(m,null,[e(s(k??"-"),1)],64))),256))]}),_:2},1024),e(),l(r,{layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.connected")),1)]),body:t(()=>[e(s(a("common.formats.datetime",{value:Date.parse(o.data.connectTime??"")})),1)]),_:2},1024),e(),o.data.disconnectTime?(n(),u(r,{key:0,layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.disconnected")),1)]),body:t(()=>[e(s(a("common.formats.datetime",{value:Date.parse(o.data.disconnectTime)})),1)]),_:2},1024)):h("",!0),e(),l(r,{layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.responses")),1)]),body:t(()=>{var i;return[(n(!0),c(m,null,b([((i=o.data.status)==null?void 0:i.total)??{}],d=>(n(),c(m,null,[e(s(d.responsesSent)+"/"+s(d.responsesAcknowledged),1)],64))),256))]}),_:2},1024),e(),o.data.zoneInstanceId?(n(),u(r,{key:1,layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.zoneInstanceId")),1)]),body:t(()=>[e(s(o.data.zoneInstanceId),1)]),_:2},1024)):h("",!0),e(),o.data.globalInstanceId?(n(),u(r,{key:2,layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.globalInstanceId")),1)]),body:t(()=>[e(s(o.data.globalInstanceId),1)]),_:2},1024)):h("",!0),e(),l(r,{layout:"horizontal"},{title:t(()=>[e(s(a("subscriptions.routes.item.headers.id")),1)]),body:t(()=>[e(s(o.data.id),1)]),_:2},1024)]),e(),p("div",C,[p("div",null,[V(w.$slots,"default")]),e(),l(r,{class:"mt-4",layout:"horizontal"},{title:t(()=>[p("strong",null,s(a("subscriptions.routes.item.headers.type")),1)]),body:t(()=>[e(s(a("subscriptions.routes.item.headers.stat")),1)]),_:2},1024),e(),(n(!0),c(m,null,b(Object.entries(((y=o.data.status)==null?void 0:y.stat)??{}),([i,d])=>(n(),u(r,{key:i,layout:"horizontal"},{title:t(()=>[e(s(i),1)]),body:t(()=>[e(s(d.responsesSent)+"/"+s(d.responsesAcknowledged),1)]),_:2},1024))),128))])]}),_:2},1024)]),_:3})}}});export{B as default};
