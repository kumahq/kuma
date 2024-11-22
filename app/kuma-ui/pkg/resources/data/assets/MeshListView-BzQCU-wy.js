import{d as z,e as n,o as C,m as A,w as e,a as t,k as p,b as l,l as k,D as V,A as b,t as m}from"./index-DpJ_igul.js";const T={class:"stack"},D=["innerHTML"],x=z({__name:"MeshListView",setup(L){return(R,r)=>{const _=n("RouteTitle"),d=n("XAction"),h=n("XActionGroup"),u=n("DataCollection"),g=n("DataLoader"),w=n("KCard"),f=n("AppView"),v=n("RouteView");return C(),A(v,{name:"mesh-list-view",params:{page:1,size:50,mesh:""}},{default:e(({route:o,t:i,me:c,uri:y})=>[t(f,{docs:i("meshes.href.docs")},{title:e(()=>[p("h1",null,[t(_,{title:i("meshes.routes.items.title")},null,8,["title"])])]),default:e(()=>[r[4]||(r[4]=l()),p("div",T,[p("div",{innerHTML:i("meshes.routes.items.intro",{},{defaultMessage:""})},null,8,D),r[3]||(r[3]=l()),t(w,null,{default:e(()=>[t(g,{src:y(k(V),"/mesh-insights",{},{page:o.params.page,size:o.params.size})},{loadable:e(({data:a})=>[t(u,{type:"meshes",items:(a==null?void 0:a.items)??[void 0],page:o.params.page,"page-size":o.params.size,total:a==null?void 0:a.total,onChange:o.update},{default:e(()=>[t(b,{class:"mesh-collection","data-testid":"mesh-collection",headers:[{...c.get("headers.name"),label:i("meshes.common.name"),key:"name"},{...c.get("headers.services"),label:i("meshes.routes.items.collection.services"),key:"services"},{...c.get("headers.dataplanes"),label:i("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{...c.get("headers.actions"),label:"Actions",key:"actions",hideLabel:!0}],items:a==null?void 0:a.items,"is-selected-row":s=>s.name===o.params.mesh,onResize:c.set},{name:e(({row:s})=>[t(d,{"data-action":"",to:{name:"mesh-detail-view",params:{mesh:s.name},query:{page:o.params.page,size:o.params.size}}},{default:e(()=>[l(m(s.name),1)]),_:2},1032,["to"])]),services:e(({row:s})=>[l(m(s.services.internal),1)]),dataplanes:e(({row:s})=>[l(m(s.dataplanesByType.standard.online)+" / "+m(s.dataplanesByType.standard.total),1)]),actions:e(({row:s})=>[t(h,null,{default:e(()=>[t(d,{to:{name:"mesh-detail-view",params:{mesh:s.name}}},{default:e(()=>[l(m(i("common.collection.actions.view")),1)]),_:2},1032,["to"])]),_:2},1024)]),_:2},1032,["headers","items","is-selected-row","onResize"])]),_:2},1032,["items","page","page-size","total","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1032,["docs"])]),_:1})}}});export{x as default};
