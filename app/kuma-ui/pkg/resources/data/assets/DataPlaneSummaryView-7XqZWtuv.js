import{d as N,e as d,o as r,m as c,w as t,a,k as l,t as s,b as e,c as u,H as g,J as f,n as X,X as p,S as A,l as k,a0 as K,K as E,p as m,$ as v,q as I}from"./index-CgC5RQPZ.js";import{q as L}from"./kong-icons.es676-DpwKnWMp.js";import{T}from"./TagList-CCBtnGQE.js";const O={class:"stack"},U={class:"stack-with-borders"},q={class:"status-with-reason"},Z={key:0},F={class:"mt-4"},G={class:"stack-with-borders"},H={class:"mt-4 stack"},J={class:"mt-2 stack-with-borders"},j=N({__name:"DataPlaneSummaryView",props:{items:{}},setup(x){const b=x;return(M,Q)=>{const C=d("XEmptyState"),D=d("RouteTitle"),w=d("XAction"),P=d("KTooltip"),y=d("DataCollection"),z=d("XBadge"),V=d("AppView"),R=d("RouteView");return r(),c(R,{name:"data-plane-summary-view",params:{dataPlane:""}},{default:t(({t:o,route:$,can:B})=>[a(y,{items:b.items,predicate:h=>h.id===$.params.dataPlane,find:!0},{empty:t(()=>[a(C,null,{title:t(()=>[l("h2",null,s(o("common.collection.summary.empty_title",{type:"Data Plane Proxy"})),1)]),default:t(()=>[e(),l("p",null,s(o("common.collection.summary.empty_message",{type:"Data Plane Proxy"})),1)]),_:2},1024)]),default:t(({items:h})=>[(r(!0),u(g,null,f([h[0]],n=>(r(),c(V,{key:n.id},{title:t(()=>[l("h2",{class:X(`type-${n.dataplaneType}`)},[a(w,{to:{name:"data-plane-detail-view",params:{dataPlane:n.id}}},{default:t(()=>[a(D,{title:o("data-planes.routes.item.title",{name:n.name})},null,8,["title"])]),_:2},1032,["to"])],2)]),default:t(()=>[e(),l("div",O,[l("div",U,[a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.status")),1)]),body:t(()=>[l("div",q,[a(A,{status:n.status},null,8,["status"]),e(),n.dataplaneType==="standard"?(r(),c(y,{key:0,items:n.dataplane.networking.inbounds,predicate:_=>_.state!=="Ready",empty:!1},{default:t(({items:_})=>[a(P,{class:"reason-tooltip"},{content:t(()=>[l("ul",null,[(r(!0),u(g,null,f(_,i=>(r(),u("li",{key:`${i.service}:${i.port}`},s(o("data-planes.routes.item.unhealthy_inbound",{service:i.service,port:i.port})),1))),128))])]),default:t(()=>[a(k(L),{color:k(K),size:k(E)},null,8,["color","size"]),e()]),_:2},1024)]),_:2},1032,["items","predicate"])):m("",!0)])]),_:2},1024),e(),a(p,{layout:"horizontal"},{title:t(()=>[e(`
                    Type
                  `)]),body:t(()=>[e(s(o(`data-planes.type.${n.dataplaneType}`)),1)]),_:2},1024),e(),n.namespace.length>0?(r(),c(p,{key:0,layout:"horizontal"},{title:t(()=>[e(s(o("data-planes.routes.item.namespace")),1)]),body:t(()=>[e(s(n.namespace),1)]),_:2},1024)):m("",!0),e(),B("use zones")&&n.zone?(r(),c(p,{key:1,layout:"horizontal"},{title:t(()=>[e(`
                    Zone
                  `)]),body:t(()=>[a(w,{to:{name:"zone-cp-detail-view",params:{zone:n.zone}}},{default:t(()=>[e(s(n.zone),1)]),_:2},1032,["to"])]),_:2},1024)):m("",!0),e(),a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("data-planes.routes.item.last_updated")),1)]),body:t(()=>[e(s(o("common.formats.datetime",{value:Date.parse(n.modificationTime)})),1)]),_:2},1024)]),e(),n.dataplane.networking.gateway?(r(),u("div",Z,[l("h3",null,s(o("data-planes.routes.item.gateway")),1),e(),l("div",F,[l("div",G,[a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.tags")),1)]),body:t(()=>[a(T,{alignment:"right",tags:n.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),e(),a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.address")),1)]),body:t(()=>[a(v,{text:`${n.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])])])):m("",!0),e(),n.dataplaneType==="standard"?(r(),c(y,{key:1,items:n.dataplane.networking.inbounds},{default:t(({items:_})=>[l("div",null,[l("h3",null,s(o("data-planes.routes.item.inbounds")),1),e(),l("div",H,[(r(!0),u(g,null,f(_,(i,S)=>(r(),u("div",{key:S,class:"inbound"},[l("h4",null,[a(v,{text:i.tags["kuma.io/service"]},{default:t(()=>[e(s(o("data-planes.routes.item.inbound_name",{service:i.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),e(),l("div",J,[a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.state")),1)]),body:t(()=>[i.state==="Ready"?(r(),c(z,{key:0,appearance:"success"},{default:t(()=>[e(s(o(`http.api.value.${i.state}`)),1)]),_:2},1024)):(r(),c(z,{key:1,appearance:"danger"},{default:t(()=>[e(s(o(`http.api.value.${i.state}`)),1)]),_:2},1024))]),_:2},1024),e(),a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.tags")),1)]),body:t(()=>[a(T,{alignment:"right",tags:i.tags},null,8,["tags"])]),_:2},1024),e(),a(p,{layout:"horizontal"},{title:t(()=>[e(s(o("http.api.property.address")),1)]),body:t(()=>[a(v,{text:i.addressPort},null,8,["text"])]),_:2},1024)])]))),128))])])]),_:2},1032,["items"])):m("",!0)])]),_:2},1024))),128))]),_:2},1032,["items","predicate"])]),_:1})}}}),et=I(j,[["__scopeId","data-v-915b988c"]]);export{et as default};
