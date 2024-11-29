import{d as D,e as m,o as r,p as u,w as e,a as o,l as d,Q as p,b as a,t as s,S as R,c as _,J as z,K as c,m as v,W as $,L as A,q as g,V as f,_ as K}from"./index-DH1Ug2X_.js";import{q as O}from"./kong-icons.es678-BjvL39h3.js";import{T}from"./TagList-BktRq3Gc.js";const S={class:"stack"},I={class:"stack-with-borders"},L={class:"status-with-reason"},P={key:0},U={class:"mt-4"},X={class:"stack-with-borders"},E={class:"mt-4 stack"},q={class:"mt-2 stack-with-borders"},Z=D({__name:"DataPlaneSummaryOverviewView",props:{data:{},routeName:{}},setup(b){const n=b;return(F,t)=>{const x=m("KTooltip"),k=m("DataCollection"),V=m("XAction"),w=m("XBadge"),C=m("AppView"),h=m("RouteView");return r(),u(h,{name:n.routeName,params:{dataPlane:""}},{default:e(({t:i,can:N})=>[o(C,null,{default:e(()=>[d("div",S,[d("div",I,[o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.status")),1)]),body:e(()=>[d("div",L,[o(R,{status:n.data.status},null,8,["status"]),t[1]||(t[1]=a()),n.data.dataplaneType==="standard"?(r(),u(k,{key:0,items:n.data.dataplane.networking.inbounds,predicate:y=>y.state!=="Ready",empty:!1},{default:e(({items:y})=>[o(x,{class:"reason-tooltip"},{content:e(()=>[d("ul",null,[(r(!0),_(z,null,c(y,l=>(r(),_("li",{key:`${l.service}:${l.port}`},s(i("data-planes.routes.item.unhealthy_inbound",{service:l.service,port:l.port})),1))),128))])]),default:e(()=>[o(v(O),{color:v($),size:v(A)},null,8,["color","size"]),t[0]||(t[0]=a())]),_:2},1024)]),_:2},1032,["items","predicate"])):g("",!0)])]),_:2},1024),t[9]||(t[9]=a()),o(p,{layout:"horizontal"},{title:e(()=>t[3]||(t[3]=[a(`
              Type
            `)])),body:e(()=>[a(s(i(`data-planes.type.${n.data.dataplaneType}`)),1)]),_:2},1024),t[10]||(t[10]=a()),n.data.namespace.length>0?(r(),u(p,{key:0,layout:"horizontal"},{title:e(()=>[a(s(i("data-planes.routes.item.namespace")),1)]),body:e(()=>[a(s(n.data.namespace),1)]),_:2},1024)):g("",!0),t[11]||(t[11]=a()),N("use zones")&&n.data.zone?(r(),u(p,{key:1,layout:"horizontal"},{title:e(()=>t[6]||(t[6]=[a(`
              Zone
            `)])),body:e(()=>[o(V,{to:{name:"zone-cp-detail-view",params:{zone:n.data.zone}}},{default:e(()=>[a(s(n.data.zone),1)]),_:1},8,["to"])]),_:1})):g("",!0),t[12]||(t[12]=a()),o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.modificationTime")),1)]),body:e(()=>[a(s(i("common.formats.datetime",{value:Date.parse(n.data.modificationTime)})),1)]),_:2},1024)]),t[24]||(t[24]=a()),n.data.dataplane.networking.gateway?(r(),_("div",P,[d("h3",null,s(i("data-planes.routes.item.gateway")),1),t[16]||(t[16]=a()),d("div",U,[d("div",X,[o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.tags")),1)]),body:e(()=>[o(T,{alignment:"right",tags:n.data.dataplane.networking.gateway.tags},null,8,["tags"])]),_:2},1024),t[15]||(t[15]=a()),o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.address")),1)]),body:e(()=>[o(f,{text:`${n.data.dataplane.networking.address}`},null,8,["text"])]),_:2},1024)])])])):g("",!0),t[25]||(t[25]=a()),n.data.dataplaneType==="standard"?(r(),u(k,{key:1,items:n.data.dataplane.networking.inbounds},{default:e(({items:y})=>[d("div",null,[d("h3",null,s(i("data-planes.routes.item.inbounds")),1),t[23]||(t[23]=a()),d("div",E,[(r(!0),_(z,null,c(y,(l,B)=>(r(),_("div",{key:B,class:"inbound"},[d("h4",null,[o(f,{text:l.tags["kuma.io/service"]},{default:e(()=>[a(s(i("data-planes.routes.item.inbound_name",{service:l.tags["kuma.io/service"]})),1)]),_:2},1032,["text"])]),t[22]||(t[22]=a()),d("div",q,[o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.state")),1)]),body:e(()=>[l.state==="Ready"?(r(),u(w,{key:0,appearance:"success"},{default:e(()=>[a(s(i(`http.api.value.${l.state}`)),1)]),_:2},1024)):(r(),u(w,{key:1,appearance:"danger"},{default:e(()=>[a(s(i(`http.api.value.${l.state}`)),1)]),_:2},1024))]),_:2},1024),t[20]||(t[20]=a()),o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.tags")),1)]),body:e(()=>[o(T,{alignment:"right",tags:l.tags},null,8,["tags"])]),_:2},1024),t[21]||(t[21]=a()),o(p,{layout:"horizontal"},{title:e(()=>[a(s(i("http.api.property.address")),1)]),body:e(()=>[o(f,{text:l.addressPort},null,8,["text"])]),_:2},1024)])]))),128))])])]),_:2},1032,["items"])):g("",!0)])]),_:2},1024)]),_:1},8,["name"])}}}),W=K(Z,[["__scopeId","data-v-5ef57640"]]);export{W as default};
