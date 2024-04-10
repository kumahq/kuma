import{d as C,l as b,a as m,o as l,b as p,w as e,e as a,f as n,t as c,p as t,k as A,x as K,m as r,E as x,y as N,c as z,F as S,z as $,q as L,_ as E}from"./index-C_L4An_H.js";import{A as D}from"./AppCollection-DyLcZM0-.js";import{S as F}from"./StatusBadge-DagklHIN.js";const Z=C({__name:"MeshInsightsList",props:{items:{}},setup(w){const{t:s}=b(),u=w;return(k,f)=>{var d;const y=m("RouterLink");return l(),p(D,{headers:[{label:t(s)("meshes.components.mesh-insights-list.name"),key:"name"},{label:t(s)("meshes.components.mesh-insights-list.services"),key:"services"},{label:t(s)("meshes.components.mesh-insights-list.dataplanes"),key:"dataplanes"}],items:u.items,total:(d=u.items)==null?void 0:d.length,"empty-state-message":t(s)("common.emptyState.message",{type:t(s)("meshes.common.type",{count:2})}),"empty-state-cta-to":t(s)("meshes.href.docs"),"empty-state-cta-text":t(s)("common.documentation")},{name:e(({row:o})=>[a(y,{to:{name:"mesh-detail-view",params:{mesh:o.name}}},{default:e(()=>[n(c(o.name),1)]),_:2},1032,["to"])]),services:e(({row:o})=>[n(c(o.services.internal),1)]),dataplanes:e(({row:o})=>[n(c(o.dataplanesByType.standard.online)+" / "+c(o.dataplanesByType.standard.total),1)]),_:1},8,["headers","items","total","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),q=C({__name:"ZoneControlPlanesList",props:{items:{}},setup(w){const{t:s}=b(),u=A(),k=w;return(f,y)=>{var o;const d=m("RouterLink");return l(),p(D,{headers:[{label:t(s)("zone-cps.components.zone-control-planes-list.name"),key:"name"},{label:t(s)("zone-cps.components.zone-control-planes-list.status"),key:"status"}],items:k.items,total:(o=k.items)==null?void 0:o.length,"empty-state-title":t(s)("zone-cps.empty_state.title"),"empty-state-message":t(u)("create zones")?t(s)("zone-cps.empty_state.message"):t(s)("common.emptyState.message",{type:"Zones"}),"empty-state-cta-to":t(u)("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":t(s)("zones.index.create")},{name:e(({row:_})=>[a(d,{to:{name:"zone-cp-detail-view",params:{zone:_.name}}},{default:e(()=>[n(c(_.name),1)]),_:2},1032,["to"])]),status:e(({row:_})=>[a(F,{status:_.state},null,8,["status"])]),_:1},8,["headers","items","total","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),I={key:2,class:"stack","data-testid":"detail-view-details"},M={class:"columns"},j={class:"card-header"},G={class:"card-title"},H={key:0,class:"card-actions"},J={class:"card-header"},O={class:"card-title"},Q=C({__name:"ControlPlaneDetailView",setup(w){const s=K();return(u,k)=>{const f=m("RouteTitle"),y=m("RouterLink"),d=m("KButton"),o=m("DataSource"),_=m("KCard"),P=m("AppView"),T=m("RouteView");return l(),p(T,{name:"home"},{default:e(({can:g,t:h})=>[a(P,null,{title:e(()=>[r("h1",null,[a(f,{title:h("main-overview.routes.item.title")},null,8,["title"])])]),default:e(()=>[n(),a(o,{src:"/global-insight"},{default:e(({data:B,error:R})=>[R?(l(),p(x,{key:0,error:R},null,8,["error"])):B===void 0?(l(),p(N,{key:1})):(l(),z("div",I,[a(t(s),{"can-use-zones":g("use zones"),"global-insight":B},null,8,["can-use-zones","global-insight"]),n(),r("div",M,[g("use zones")?(l(),p(_,{key:0},{default:e(()=>[a(o,{src:"/zone-cps?page=1&size=10"},{default:e(({data:i,error:v})=>{var V;return[v?(l(),p(x,{key:0,error:v},null,8,["error"])):(l(),z(S,{key:1},[r("div",j,[r("div",G,[r("h2",null,c(h("main-overview.detail.zone_control_planes.title")),1),n(),a(y,{to:{name:"zone-cp-list-view"}},{default:e(()=>[n(c(h("main-overview.detail.health.view_all")),1)]),_:2},1024)]),n(),g("create zones")&&(((V=i==null?void 0:i.items)==null?void 0:V.length)??!1)?(l(),z("div",H,[a(d,{appearance:"primary",to:{name:"zone-create-view"}},{default:e(()=>[a(t($)),n(" "+c(h("zones.index.create")),1)]),_:2},1024)])):L("",!0)]),n(),a(q,{"data-testid":"zone-control-planes-details",items:i==null?void 0:i.items},null,8,["items"])],64))]}),_:2},1024)]),_:2},1024)):L("",!0),n(),a(_,null,{default:e(()=>[a(o,{src:"/mesh-insights?page=1&size=10"},{default:e(({data:i,error:v})=>[v?(l(),p(x,{key:0,error:v},null,8,["error"])):(l(),z(S,{key:1},[r("div",J,[r("div",O,[r("h2",null,c(h("main-overview.detail.meshes.title")),1),n(),a(y,{to:{name:"mesh-list-view"}},{default:e(()=>[n(c(h("main-overview.detail.health.view_all")),1)]),_:2},1024)])]),n(),a(Z,{"data-testid":"meshes-details",items:i==null?void 0:i.items},null,8,["items"])],64))]),_:2},1024)]),_:2},1024)])]))]),_:2},1024)]),_:2},1024)]),_:1})}}}),Y=E(Q,[["__scopeId","data-v-19c46594"]]);export{Y as default};
