import{d as K,r as p,o as r,m as c,w as t,b as e,ac as N,e as a,t as o,k as l,c as u,F as g,G as k,n as $,U as d,S as L,l as f,V as S,K as I,p as m,T as w,q as U}from"./index-Is4zmHdk.js";import{m as A}from"./kong-icons.es338-BhuQ6P9U.js";import{T}from"./TagList-BpgBN_GZ.js";const O={class:"stack"},E={class:"stack-with-borders"},F={class:"status-with-reason"},G={key:0},q={class:"mt-4"},Z={class:"stack-with-borders"},j={class:"mt-4 stack"},H={class:"mt-2 stack-with-borders"},J=K({__name:"DataPlaneSummaryView",props:{items:{}},setup(b){const x=b;return(M,Q)=>{const z=p("RouteTitle"),C=p("RouterLink"),V=p("KTooltip"),y=p("DataCollection"),v=p("KBadge"),D=p("AppView"),P=p("RouteView");return r(),c(P,{name:"data-plane-summary-view",params:{dataPlane:""}},{default:t(({t:s,route:R})=>[e(y,{items:x.items,predicate:h=>h.id===R.params.dataPlane,find:!0},{empty:t(()=>[e(N,null,{title:t(()=>[a(o(s("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:t(()=>[a(),l("p",null,o(s("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:t(({items:h})=>[(r(!0),u(g,null,k([h[0]],n=>(r(),c(D,{key:n.id},{title:t(()=>[l("h2",{class:$(`type-${n.dataplaneType}`)},[e(C,{to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:t(()=>[e(z,{title:s("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:t(()=>[a(),l("div",O,[l("div",E,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.status")),1)]),body:t(()=>[l("div",F,[e(L,{status:n.status},null,8,["status"]),a(),n.dataplaneType==="standard"?(r(),c(y,{key:0,items:n.dataplane.networking.inbounds,predicate:_=>!_.health.ready,empty:!1},{default:t(({items:_})=>[e(V,{class:"reason-tooltip"},{content:t(()=>[l("ul",null,[(r(!0),u(g,null,k(_,i=>(r(),u("li",{key:`${i.service}:${i.port}`},o(s("data-planes.routes.item.unhealthy_inbound",{service:i.service,port:i.port})),1))),128))])]),default:t(()=>[e(f(A),{color:f(S),size:f(I)},null,8,["color","size"]),a()]),_:2},1024)]),_:2},1032,["items","predicate"])):m("",!0)])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(`
                    Type
                  `)]),body:t(()=>[a(o(s(`data-planes.type.${n.dataplaneType}`)),1)]),_:2},1024),a(),n.namespace.length>0?(r(),c(d,{key:0,layout:"horizontal"},{title:t(()=>[a(o(s("data-planes.routes.item.namespace")),1)]),body:t(()=>[a(o(n.namespace),1)]),_:2},1024)):m("",!0),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("data-planes.routes.item.last_updated")),1)]),body:t(()=>[a(o(s("common.formats.datetime",{value:Date.parse(n.modificationTime)})),1)]),_:2},1024)]),a(),n.dataplane.networking.gateway?(r(),u("div",G,[l("h3",null,o(s("data-planes.routes.item.gateway")),1),a(),l("div",q,[l("div",Z,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.tags")),1)]),body:t(()=>[e(T,{alignment:"right",tags:n.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.address")),1)]),body:t(()=>[e(w,{text:`${n.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])])])):m("",!0),a(),n.dataplaneType==="standard"?(r(),c(y,{key:1,items:n.dataplane.networking.inbounds},{default:t(({items:_})=>[l("div",null,[l("h3",null,o(s("data-planes.routes.item.inbounds")),1),a(),l("div",j,[(r(!0),u(g,null,k(_,(i,B)=>(r(),u("div",{key:B,class:"inbound"},[l("h4",null,[e(w,{text:i.tags["kuma.io/service"]},{default:t(()=>[a(o(s("data-planes.routes.item.inbound_name",{service:i.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),a(),l("div",H,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.status")),1)]),body:t(()=>[i.health.ready?(r(),c(v,{key:0,appearance:"success"},{default:t(()=>[a(o(s("data-planes.routes.item.health.ready")),1)]),_:2},1024)):(r(),c(v,{key:1,appearance:"danger"},{default:t(()=>[a(o(s("data-planes.routes.item.health.not_ready")),1)]),_:2},1024))]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.tags")),1)]),body:t(()=>[e(T,{alignment:"right",tags:i.tags},null,8,["tags"])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.address")),1)]),body:t(()=>[e(w,{text:i.addressPort},null,8,["text"])]),_:2},1024)])]))),128))])])]),_:2},1032,["items"])):m("",!0)])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),tt=U(J,[["__scopeId","data-v-4eb4457e"]]);export{tt as default};
