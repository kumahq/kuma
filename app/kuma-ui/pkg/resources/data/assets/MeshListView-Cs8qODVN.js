import{d as y,e as n,o as C,m as z,w as e,a as t,k as r,b as l,l as A,C as k,A as V,t as m}from"./index-wlZh7JTj.js";const b={class:"stack"},T=["innerHTML"],x=y({__name:"MeshListView",setup(L){return(R,B)=>{const d=n("RouteTitle"),p=n("XAction"),_=n("XActionGroup"),h=n("DataCollection"),u=n("DataLoader"),g=n("KCard"),w=n("AppView"),f=n("RouteView");return C(),z(f,{name:"mesh-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:o,t:i,me:c,uri:v})=>[t(w,{docs:i("meshes.href.docs")},{title:e(()=>[r("h1",null,[t(d,{title:i("meshes.routes.items.title")},null,8,["title"])])]),default:e(()=>[l(),r("div",b,[r("div",{innerHTML:i("meshes.routes.items.intro",{},{defaultMessage:""})},null,8,T),l(),t(g,null,{default:e(()=>[t(u,{src:v(A(k),"/mesh-insights",{},{page:o.params.page,size:o.params.size})},{loadable:e(({data:a})=>[t(h,{type:"meshes",items:(a==null?void 0:a.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,onChange:o.update},{default:e(()=>[t(V,{class:"mesh-collection","data-testid":"mesh-collection",headers:[{...c.get("headers.name"),label:i("meshes.common.name"),key:"name"},{...c.get("headers.services"),label:i("meshes.routes.items.collection.services"),key:"services"},{...c.get("headers.dataplanes"),label:i("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,"is-selected-row":s=>s.name===o.params.mesh,onResize:c.set},{name:e(({row:s})=>[t(p,{"data-action":"",to:{name:"mesh-detail-view",params:{mesh:s.name},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[l(m(s.name),1)]),_:2},1032,["to"])]),services:e(({row:s})=>[l(m(s.services.internal),1)]),dataplanes:e(({row:s})=>[l(m(s.dataplanesByType.standard.online)+" / "+m(s.dataplanesByType.standard.total),1)]),actions:e(({row:s})=>[t(_,null,{default:e(()=>[t(p,{to:{name:"mesh-detail-view",params:{mesh:s.name}}},{default:e(()=>[l(m(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1032,["docs"])]),_:1})}}});export{x as default};
