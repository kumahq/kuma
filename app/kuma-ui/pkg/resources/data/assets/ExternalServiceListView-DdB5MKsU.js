import{d as A,r as o,o as c,p as d,w as e,b as t,e as l,m as C,at as y,A as V,V as u,t as m,c as b,J as R}from"./index-yoi81zLz.js";const B=A({__name:"ExternalServiceListView",setup(D){return(L,p)=>{const g=o("RouteTitle"),_=o("XAction"),h=o("XActionGroup"),v=o("DataCollection"),x=o("DataLoader"),f=o("KCard"),w=o("AppView"),k=o("RouteView");return c(),d(k,{name:"external-service-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:n,t:i,me:r,uri:z})=>[t(g,{render:!1,title:i("external-services.routes.items.title")},null,8,["title"]),p[2]||(p[2]=l()),t(w,{docs:i("external-services.href.docs")},{default:e(()=>[t(f,null,{default:e(()=>[t(x,{src:z(C(y),"/meshes/:mesh/external-services",{mesh:n.params.mesh},{page:n.params.page,size:n.params.size})},{loadable:e(({data:a})=>[t(v,{type:"external-services",items:(a==null?void 0:a.items)??[void 0],page:n.params.page,"page-size":n.params.size,total:a==null?void 0:a.total,onChange:n.update},{default:e(()=>[t(V,{class:"external-service-collection","data-testid":"external-service-collection",headers:[{...r.get("headers.name"),label:"Name",key:"name"},{...r.get("headers.address"),label:"Address",key:"address"},{...r.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,onResize:r.set},{name:e(({row:s})=>[t(u,{text:s.name},{default:e(()=>[t(_,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name},query:{page:n.params.page,size:n.params.size}}},{default:e(()=>[l(m(s.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),address:e(({row:s})=>[s.networking.address?(c(),d(u,{key:0,text:s.networking.address},null,8,["text"])):(c(),b(R,{key:1},[l(m(i("common.collection.none")),1)],64))]),actions:e(({row:s})=>[t(h,null,{default:e(()=>[t(_,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name}}},{default:e(()=>[l(m(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{B as default};
