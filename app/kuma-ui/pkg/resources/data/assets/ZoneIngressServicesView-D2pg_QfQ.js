import{p as h}from"./kong-icons.es337-C4tX_a7c.js";import{d as g,h as s,o as k,a as f,w as e,j as t,g as b,k as a,C as y,t as i,z as C}from"./index-9gITI0JG.js";const B=g({__name:"ZoneIngressServicesView",props:{data:{}},setup(r){const l=r;return(K,V)=>{const m=s("RouteTitle"),c=s("RouterLink"),p=s("KButton"),u=s("KDropdownItem"),d=s("KDropdown"),_=s("KCard"),v=s("AppView"),w=s("RouteView");return k(),f(w,{name:"zone-ingress-services-view"},{default:e(({t:n})=>[t(v,null,{title:e(()=>[b("h2",null,[t(m,{title:n("zone-ingresses.routes.item.navigation.zone-ingress-services-view")},null,8,["title"])])]),default:e(()=>[a(),t(_,null,{default:e(()=>[t(y,{"data-testid":"available-services-collection","empty-state-message":n("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Protocol",key:"protocol"},{label:"No. Instances",key:"instances"},{label:"Actions",key:"actions",hideLabel:!0}],items:l.data.zoneIngress.availableServices},{name:e(({row:o})=>[t(c,{to:{name:"service-detail-view",params:{mesh:o.mesh,service:o.tags["kuma.io/service"]}}},{default:e(()=>[a(i(o.tags["kuma.io/service"]),1)]),_:2},1032,["to"])]),mesh:e(({row:o})=>[t(c,{to:{name:"mesh-detail-view",params:{mesh:o.mesh}}},{default:e(()=>[a(i(o.mesh),1)]),_:2},1032,["to"])]),protocol:e(({row:o})=>[a(i(o.tags["kuma.io/protocol"]??n("common.collection.none")),1)]),instances:e(({row:o})=>[a(i(o.instances),1)]),actions:e(({row:o})=>[t(d,{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(p,{class:"non-visual-button",appearance:"secondary",icon:""},{default:e(()=>[t(C(h))]),_:1})]),items:e(()=>[t(u,{item:{to:{name:"service-detail-view",params:{mesh:o.mesh,service:o.tags["kuma.io/service"]}},label:n("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["empty-state-message","items"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{B as default};
