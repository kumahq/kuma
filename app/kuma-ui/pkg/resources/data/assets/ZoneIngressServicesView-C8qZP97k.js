import{d as w,r as s,o as h,m as g,w as o,b as t,e as a,A as k,t as n}from"./index-DOjcqG3h.js";const A=w({__name:"ZoneIngressServicesView",props:{data:{}},setup(r){const l=r;return(f,y)=>{const m=s("RouteTitle"),c=s("RouterLink"),p=s("XAction"),u=s("XActionGroup"),_=s("KCard"),d=s("AppView"),v=s("RouteView");return h(),g(v,{name:"zone-ingress-services-view"},{default:o(({t:i})=>[t(m,{render:!1,title:i("zone-ingresses.routes.item.navigation.zone-ingress-services-view")},null,8,["title"]),a(),t(d,null,{default:o(()=>[t(_,null,{default:o(()=>[t(k,{"data-testid":"available-services-collection","empty-state-message":i("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Protocol",key:"protocol"},{label:"No. Instances",key:"instances"},{label:"Actions",key:"actions",hideLabel:!0}],items:l.data.zoneIngress.availableServices},{name:o(({row:e})=>[t(c,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.tags["kuma.io/service"]}}},{default:o(()=>[a(n(e.tags["kuma.io/service"]),1)]),_:2},1032,["to"])]),mesh:o(({row:e})=>[t(c,{to:{name:"mesh-detail-view",params:{mesh:e.mesh}}},{default:o(()=>[a(n(e.mesh),1)]),_:2},1032,["to"])]),protocol:o(({row:e})=>[a(n(e.tags["kuma.io/protocol"]??i("common.collection.none")),1)]),instances:o(({row:e})=>[a(n(e.instances),1)]),actions:o(({row:e})=>[t(u,null,{default:o(()=>[t(p,{to:{name:"service-detail-view",params:{mesh:e.mesh,service:e.tags["kuma.io/service"]}}},{default:o(()=>[a(n(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["empty-state-message","items"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{A as default};
