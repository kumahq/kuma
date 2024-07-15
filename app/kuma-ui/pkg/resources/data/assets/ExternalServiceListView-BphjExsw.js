import{d as A,r as n,o as l,m as p,w as e,b as t,e as i,l as C,aH as b,A as y,T as _,t as m,c as R,F as V}from"./index-DxrN05KS.js";const B=A({__name:"ExternalServiceListView",setup(L){return(D,T)=>{const d=n("RouteTitle"),u=n("RouterLink"),h=n("XAction"),g=n("XActionGroup"),v=n("DataCollection"),x=n("DataLoader"),f=n("KCard"),w=n("AppView"),k=n("RouteView");return l(),p(k,{name:"external-service-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:o,t:r,me:c,uri:z})=>[t(d,{render:!1,title:r("external-services.routes.items.title")},null,8,["title"]),i(),t(w,{docs:r("external-services.href.docs")},{default:e(()=>[t(f,null,{default:e(()=>[t(x,{src:z(C(b),"/meshes/:mesh/external-services",{mesh:o.params.mesh},{page:o.params.page,size:o.params.size})},{loadable:e(({data:a})=>[t(v,{type:"external-services",items:(a==null?void 0:a.items)??[void 0]},{default:e(()=>[t(y,{class:"external-service-collection","data-testid":"external-service-collection",headers:[{...c.get("headers.name"),label:"Name",key:"name"},{...c.get("headers.address"),label:"Address",key:"address"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],"page-number":o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,items:a==null?void 0:a.items,onChange:o.update,onResize:c.set},{name:e(({row:s})=>[t(_,{text:s.name},{default:e(()=>[t(u,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[i(m(s.name),1)]),_:2},1032,["to"])]),_:2},1032,["text"])]),address:e(({row:s})=>[s.networking.address?(l(),p(_,{key:0,text:s.networking.address},null,8,["text"])):(l(),R(V,{key:1},[i(m(r("common.collection.none")),1)],64))]),actions:e(({row:s})=>[t(g,null,{default:e(()=>[t(h,{to:{name:"external-service-detail-view",params:{mesh:s.mesh,service:s.name}}},{default:e(()=>[i(m(r("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","page-number","page-size","total","items","onChange","onResize"])]),_:2},1032,["items"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{B as default};
