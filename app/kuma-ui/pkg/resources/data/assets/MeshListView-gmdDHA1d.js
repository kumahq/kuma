import{d as C,e as o,o as A,p as V,w as e,a,l as d,b as i,m as T,D as b,A as k,t as c}from"./index-CFsM3b-2.js";const D=["innerHTML"],B=C({__name:"MeshListView",setup(L){return(R,m)=>{const _=o("RouteTitle"),r=o("XAction"),h=o("XActionGroup"),u=o("DataCollection"),g=o("DataLoader"),w=o("XCard"),f=o("AppView"),v=o("RouteView");return A(),V(v,{name:"mesh-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:t,t:n,me:l,uri:y})=>[a(f,{docs:n("meshes.href.docs")},{title:e(()=>[d("h1",null,[a(_,{title:n("meshes.routes.items.title")},null,8,["title"])])]),default:e(()=>[m[3]||(m[3]=i()),d("div",{innerHTML:n("meshes.routes.items.intro",{},{defaultMessage:""})},null,8,D),m[4]||(m[4]=i()),a(w,null,{default:e(()=>[a(g,{variant:"list",src:y(T(b),"/mesh-insights",{},{page:t.params.page,size:t.params.size})},{default:e(({data:p})=>[a(u,{type:"meshes",items:p.items,page:t.params.page,"page-size":t.params.size,total:p.total,onChange:t.update},{default:e(({items:z})=>[a(k,{class:"mesh-collection","data-testid":"mesh-collection",headers:[{...l.get("headers.name"),label:n("meshes.common.name"),key:"name"},{...l.get("headers.services"),label:n("meshes.routes.items.collection.services"),key:"services"},{...l.get("headers.dataplanes"),label:n("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{...l.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:z,"is-selected-row":s=>s.name===t.params.mesh,onResize:l.set},{name:e(({row:s})=>[a(r,{"data-action":"",to:{name:"mesh-detail-view",params:{mesh:s.name},query:{page:t.params.page,size:t.params.size}}},{default:e(()=>[i(c(s.name),1)]),_:2},1032,["to"])]),services:e(({row:s})=>[i(c(s.services.internal),1)]),dataplanes:e(({row:s})=>[i(c(s.dataplanesByType.standard.online)+" / "+c(s.dataplanesByType.standard.total),1)]),actions:e(({row:s})=>[a(h,null,{default:e(()=>[a(r,{to:{name:"mesh-detail-view",params:{mesh:s.name}}},{default:e(()=>[i(c(n("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)]),_:2},1032,["docs"])]),_:1})}}});export{B as default};
