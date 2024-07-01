import{d as z,y as A,h as l,o as v,b as D,j as e,w as s,C as T,z as i,k as t,t as c,D as x,E as B,a as C,g as o,G as P,H as X,e as L,J as S,A as N}from"./index-9gITI0JG.js";const R=z({__name:"MeshInsightsList",props:{items:{}},setup(w){const{t:r}=A(),m=w;return(y,V)=>{const h=l("XAction"),_=l("DataCollection");return v(),D("div",null,[e(_,{items:m.items??[void 0],type:"meshes"},{default:s(()=>{var d;return[e(T,{headers:[{label:i(r)("meshes.components.mesh-insights-list.name"),key:"name"},{label:i(r)("meshes.components.mesh-insights-list.services"),key:"services"},{label:i(r)("meshes.components.mesh-insights-list.dataplanes"),key:"dataplanes"}],items:m.items,total:(d=m.items)==null?void 0:d.length},{name:s(({row:n})=>[e(h,{to:{name:"mesh-detail-view",params:{mesh:n.name}}},{default:s(()=>[t(c(n.name),1)]),_:2},1032,["to"])]),services:s(({row:n})=>[t(c(n.services.internal),1)]),dataplanes:s(({row:n})=>[t(c(n.dataplanesByType.standard.online)+" / "+c(n.dataplanesByType.standard.total),1)]),_:1},8,["headers","items","total"])]}),_:1},8,["items"])])}}}),I={class:"stack"},$={class:"columns"},E={class:"card-header"},K={class:"card-title"},Z={class:"card-actions"},j={class:"card-header"},G={class:"card-title"},H=z({__name:"ControlPlaneDetailView",setup(w){const r=x(),m=B();return(y,V)=>{const h=l("RouteTitle"),_=l("DataLoader"),d=l("XAction"),n=l("XTeleportSlot"),f=l("KCard"),b=l("AppView"),k=l("RouteView");return v(),C(k,{name:"home"},{default:s(({can:g,t:p,uri:u})=>[e(b,null,{title:s(()=>[o("h1",null,[e(h,{title:p("main-overview.routes.item.title")},null,8,["title"])])]),default:s(()=>[t(),o("div",I,[e(_,{src:u(i(P),"/global-insight",{})},{default:s(({data:a})=>[e(i(r),{"can-use-zones":g("use zones"),"global-insight":a},null,8,["can-use-zones","global-insight"])]),_:2},1032,["src"]),t(),o("div",$,[g("use zones")?(v(),C(f,{key:0},{default:s(()=>[e(_,{src:u(i(X),"/zone-cps",{},{page:1,size:10})},{loadable:s(({data:a})=>[o("div",E,[o("div",K,[o("h2",null,c(p("main-overview.detail.zone_control_planes.title")),1),t(),e(d,{to:{name:"zone-cp-list-view"}},{default:s(()=>[t(c(p("main-overview.detail.health.view_all")),1)]),_:2},1024)]),t(),o("div",Z,[e(n,{name:"control-plane-detail-view-zone-actions"})])]),t(),e(i(m),{"data-testid":"zone-control-planes-details",items:a==null?void 0:a.items},null,8,["items"])]),_:2},1032,["src"])]),_:2},1024)):L("",!0),t(),e(f,null,{default:s(()=>[e(_,{src:u(i(S),"/mesh-insights",{},{page:1,size:10})},{loadable:s(({data:a})=>[o("div",j,[o("div",G,[o("h2",null,c(p("main-overview.detail.meshes.title")),1),t(),e(d,{to:{name:"mesh-list-view"}},{default:s(()=>[t(c(p("main-overview.detail.health.view_all")),1)]),_:2},1024)])]),t(),e(R,{"data-testid":"meshes-details",items:a==null?void 0:a.items},null,8,["items"])]),_:2},1032,["src"])]),_:2},1024)])])]),_:2},1024)]),_:1})}}}),M=N(H,[["__scopeId","data-v-2568ae39"]]);export{M as default};
