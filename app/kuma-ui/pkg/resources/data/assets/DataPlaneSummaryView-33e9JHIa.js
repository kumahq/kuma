import{d as K,r as p,o as r,m as c,w as t,b as e,ad as N,e as a,t as o,k as l,c as u,F as g,s as f,n as $,Y as d,S as L,l as k,Z as S,K as I,p as m,U as w,q as U}from"./index-BRBWbknO.js";import{m as A}from"./kong-icons.es321-CF_c0uab.js";import{T}from"./TagList-BFsDXRWm.js";const O={class:"stack"},E={class:"stack-with-borders"},F={class:"status-with-reason"},Z={key:0},q={class:"mt-4"},G={class:"stack-with-borders"},Y={class:"mt-4 stack"},j={class:"mt-2 stack-with-borders"},H=K({__name:"DataPlaneSummaryView",props:{items:{}},setup(b){const x=b;return(J,M)=>{const z=p("RouteTitle"),C=p("RouterLink"),D=p("KTooltip"),y=p("DataCollection"),v=p("KBadge"),P=p("AppView"),V=p("RouteView");return r(),c(V,{name:"data-plane-summary-view",params:{dataPlane:""}},{default:t(({t:s,route:R})=>[e(y,{items:x.items,predicate:h=>h.id===R.params.dataPlane,find:!0},{empty:t(()=>[e(N,null,{title:t(()=>[a(o(s("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:t(()=>[a(),l("p",null,o(s("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:t(({items:h})=>[(r(!0),u(g,null,f([h[0]],n=>(r(),c(P,{key:n.id},{title:t(()=>[l("h2",{class:$(`type-${n.dataplaneType}`)},[e(C,{to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:t(()=>[e(z,{title:s("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:t(()=>[a(),l("div",O,[l("div",E,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.status")),1)]),body:t(()=>[l("div",F,[e(L,{status:n.status},null,8,["status"]),a(),n.dataplaneType==="standard"?(r(),c(y,{key:0,items:n.dataplane.networking.inbounds,predicate:_=>!_.health.ready,empty:!1},{default:t(({items:_})=>[e(D,{class:"reason-tooltip"},{content:t(()=>[l("ul",null,[(r(!0),u(g,null,f(_,i=>(r(),u("li",{key:`${i.service}:${i.port}`},o(s("data-planes.routes.item.unhealthy_inbound",{service:i.service,port:i.port})),1))),128))])]),default:t(()=>[e(k(A),{color:k(S),size:k(I)},null,8,["color","size"]),a()]),_:2},1024)]),_:2},1032,["items","predicate"])):m("",!0)])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(`
                    Type
                  `)]),body:t(()=>[a(o(s(`data-planes.type.${n.dataplaneType}`)),1)]),_:2},1024),a(),n.namespace.length>0?(r(),c(d,{key:0,layout:"horizontal"},{title:t(()=>[a(o(s("data-planes.routes.item.namespace")),1)]),body:t(()=>[a(o(n.namespace),1)]),_:2},1024)):m("",!0),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("data-planes.routes.item.last_updated")),1)]),body:t(()=>[a(o(s("common.formats.datetime",{value:Date.parse(n.modificationTime)})),1)]),_:2},1024)]),a(),n.dataplane.networking.gateway?(r(),u("div",Z,[l("h3",null,o(s("data-planes.routes.item.gateway")),1),a(),l("div",q,[l("div",G,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.tags")),1)]),body:t(()=>[e(T,{alignment:"right",tags:n.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.address")),1)]),body:t(()=>[e(w,{text:`${n.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])])])):m("",!0),a(),n.dataplaneType==="standard"?(r(),c(y,{key:1,items:n.dataplane.networking.inbounds},{default:t(({items:_})=>[l("div",null,[l("h3",null,o(s("data-planes.routes.item.inbounds")),1),a(),l("div",Y,[(r(!0),u(g,null,f(_,(i,B)=>(r(),u("div",{key:B,class:"inbound"},[l("h4",null,[e(w,{text:i.tags["kuma.io/service"]},{default:t(()=>[a(o(s("data-planes.routes.item.inbound_name",{service:i.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),a(),l("div",j,[e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.status")),1)]),body:t(()=>[i.health.ready?(r(),c(v,{key:0,appearance:"success"},{default:t(()=>[a(o(s("data-planes.routes.item.health.ready")),1)]),_:2},1024)):(r(),c(v,{key:1,appearance:"danger"},{default:t(()=>[a(o(s("data-planes.routes.item.health.not_ready")),1)]),_:2},1024))]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.tags")),1)]),body:t(()=>[e(T,{alignment:"right",tags:i.tags},null,8,["tags"])]),_:2},1024),a(),e(d,{layout:"horizontal"},{title:t(()=>[a(o(s("http.api.property.address")),1)]),body:t(()=>[e(w,{text:i.addressPort},null,8,["text"])]),_:2},1024)])]))),128))])])]),_:2},1032,["items"])):m("",!0)])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),tt=U(H,[["__scopeId","data-v-9c28f6bf"]]);export{tt as default};
