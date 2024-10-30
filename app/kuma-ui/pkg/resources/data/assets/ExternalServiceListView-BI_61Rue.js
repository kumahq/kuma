import{d as z,e as o,o as r,m as _,w as e,a,b as l,l as A,aq as C,A as y,$ as d,t as m,c as V,H as b}from"./index-BGYhp_E8.js";const B=z({__name:"ExternalServiceListView",setup(R){return(D,L)=>{const u=o("RouteTitle"),p=o("XAction"),h=o("XActionGroup"),g=o("DataCollection"),v=o("DataLoader"),x=o("KCard"),f=o("AppView"),w=o("RouteView");return r(),_(w,{name:"external-service-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:n,t:c,me:i,uri:k})=>[a(u,{render:!1,title:c("external-services.routes.items.title")},null,8,["title"]),l(),a(f,{docs:c("external-services.href.docs")},{default:e(()=>[a(x,null,{default:e(()=>[a(v,{src:k(A(C),"/meshes/:mesh/external-services",{mesh:n.params.mesh},{page:n.params.page,size:n.params.size})},{loadable:e(({data:t})=>[a(g,{type:"external-services",items:(t==null?void 0:t.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:t==null?void 0:t.total,onChange:n.update},{default:e(()=>[a(y,{class:"external-service-collection","data-testid":"external-service-collection",headers:[{...i.get("headers.name"),label:"Name",key:"name"},{...i.get("headers.address"),label:"Address",key:"address"},{...i.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:t==null?void 0:t.items,onResize:i.set},{name:e(({row:s})=>[a(d,{text:s.name},{default:e(()=>[a(p,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name},query:{page:n.params.page,size:n.params.size}}},{default:e(()=>[l(m(s.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),address:e(({row:s})=>[s.networking.address?(r(),_(d,{key:0,text:s.networking.address},null,8,["text"])):(r(),V(b,{key:1},[l(m(c("common.collection.none")),1)],64))]),actions:e(({row:s})=>[a(h,null,{default:e(()=>[a(p,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name}}},{default:e(()=>[l(m(c("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{B as default};
