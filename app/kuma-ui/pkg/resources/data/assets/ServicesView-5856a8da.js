import{d as k,r as s,o as b,i as g,w as e,j as t,p as y,n,a0 as f,H as i,k as K,K as I}from"./index-b94d59a3.js";const N=k({__name:"ServicesView",props:{data:{}},setup(r){const l=r;return(C,V)=>{const m=s("RouteTitle"),c=s("RouterLink"),p=s("MoreIcon"),u=s("KButton"),d=s("KDropdownItem"),_=s("KDropdownMenu"),v=s("KCard"),w=s("AppView"),h=s("RouteView");return b(),g(h,{name:"zone-ingress-services-view"},{default:e(({t:a})=>[t(w,null,{title:e(()=>[y("h2",null,[t(m,{title:a("zone-ingresses.routes.item.navigation.zone-ingress-services-view"),render:!0},null,8,["title"])])]),default:e(()=>[n(),t(v,null,{body:e(()=>[t(f,{"data-testid":"available-services-collection","empty-state-message":a("common.emptyState.message",{type:"Services"}),headers:[{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Protocol",key:"protocol"},{label:"No. Instances",key:"instances"},{label:"Actions",key:"actions",hideLabel:!0}],items:l.data.zoneIngress.availableServices},{name:e(({row:o})=>[t(c,{to:{name:"service-detail-view",params:{mesh:o.mesh,service:o.tags["kuma.io/service"]}}},{default:e(()=>[n(i(o.tags["kuma.io/service"]),1)]),_:2},1032,["to"])]),mesh:e(({row:o})=>[t(c,{to:{name:"mesh-detail-view",params:{mesh:o.mesh}}},{default:e(()=>[n(i(o.mesh),1)]),_:2},1032,["to"])]),protocol:e(({row:o})=>[n(i(o.tags["kuma.io/protocol"]??a("common.collection.none")),1)]),instances:e(({row:o})=>[n(i(o.instances),1)]),actions:e(({row:o})=>[t(_,{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:e(()=>[t(u,{class:"non-visual-button",appearance:"secondary",size:"small"},{default:e(()=>[t(p,{size:K(I)},null,8,["size"])]),_:1})]),items:e(()=>[t(d,{item:{to:{name:"service-detail-view",params:{mesh:o.mesh,service:o.tags["kuma.io/service"]}},label:a("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:2},1032,["empty-state-message","headers","items"])]),_:2},1024)]),_:2},1024)]),_:1})}}});export{N as default};
