import{d as B,g as L,r as m,o as a,i as d,w as s,j as o,n,H as l,k as e,f as K,l as y,F as z,I as M,Y as O,p as r,Z as N,K as T,m as R,t as Z}from"./index-f09cca58.js";import{A as S}from"./AppCollection-4b4f9dc8.js";import{S as A}from"./StatusBadge-3b00ac53.js";import{g as E}from"./index-d110ad1b.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-bb9bf655.js";const D=B({__name:"MeshInsightsList",props:{items:{}},setup(g){const{t}=L(),u=g;return(f,x)=>{var p;const w=m("RouterLink");return a(),d(S,{headers:[{label:e(t)("meshes.components.mesh-insights-list.name"),key:"name"},{label:e(t)("meshes.components.mesh-insights-list.services"),key:"services"},{label:e(t)("meshes.components.mesh-insights-list.dataplanes"),key:"dataplanes"}],items:u.items,total:(p=u.items)==null?void 0:p.length,"empty-state-message":e(t)("common.emptyState.message",{type:e(t)("meshes.common.type",{count:2})}),"empty-state-cta-to":e(t)("meshes.href.docs"),"empty-state-cta-text":e(t)("common.documentation")},{name:s(({row:i})=>[o(w,{to:{name:"mesh-detail-view",params:{mesh:i.name}}},{default:s(()=>[n(l(i.name),1)]),_:2},1032,["to"])]),services:s(({row:i})=>[n(l(i.services.internal??"0"),1)]),dataplanes:s(({row:i})=>[n(l(i.dataplanesByType.standard.online??"0")+" / "+l(i.dataplanesByType.standard.total??"0"),1)]),_:1},8,["headers","items","total","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),F=B({__name:"ZoneControlPlanesList",props:{items:{}},setup(g){const{t}=L(),u=K(),f=g;return(x,w)=>{var i;const p=m("RouterLink");return a(),d(S,{headers:[{label:e(t)("zone-cps.components.zone-control-planes-list.name"),key:"name"},{label:e(t)("zone-cps.components.zone-control-planes-list.status"),key:"status"}],items:f.items,total:(i=f.items)==null?void 0:i.length,"empty-state-title":e(t)("zone-cps.empty_state.title"),"empty-state-message":e(u)("create zones")?e(t)("zone-cps.empty_state.message"):e(t)("common.emptyState.message",{type:"Zones"}),"empty-state-cta-to":e(u)("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":e(t)("zones.index.create")},{name:s(({row:_})=>[o(p,{to:{name:"zone-cp-detail-view",params:{zone:_.name}}},{default:s(()=>[n(l(_.name),1)]),_:2},1032,["to"])]),status:s(({row:_})=>[(a(!0),y(z,null,M([e(E)(_)],h=>(a(),y(z,{key:h},[h?(a(),d(A,{key:0,status:h},null,8,["status"])):(a(),y(z,{key:1},[n(l(e(t)("common.collection.none")),1)],64))],64))),128))]),_:1},8,["headers","items","total","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),P={class:"stack","data-testid":"detail-view-details"},$={class:"columns"},j={class:"card-header"},H={class:"card-title"},U={key:0,class:"card-actions"},Y={class:"card-header"},q={class:"card-title"},G=B({__name:"MainOverviewView",setup(g){const t=O();return(u,f)=>{const x=m("RouteTitle"),w=m("ErrorBlock"),p=m("RouterLink"),i=m("KButton"),_=m("DataSource"),h=m("KCard"),b=m("AppView"),I=m("RouteView");return a(),d(I,{name:"home"},{default:s(({can:C,t:v})=>[o(b,null,{title:s(()=>[r("h1",null,[o(x,{title:v("main-overview.routes.item.title"),render:!0},null,8,["title"])])]),default:s(()=>[n(),r("div",P,[o(e(t)),n(),r("div",$,[C("use zones")?(a(),d(h,{key:0},{body:s(()=>[o(_,{src:"/zone-cps?page=1&size=10"},{default:s(({data:c,error:k})=>{var V;return[k?(a(),d(w,{key:0,error:k},null,8,["error"])):(a(),y(z,{key:1},[r("div",j,[r("div",H,[r("h2",null,l(v("main-overview.detail.zone_control_planes.title")),1),n(),o(p,{to:{name:"zone-cp-list-view"}},{default:s(()=>[n(l(v("main-overview.detail.health.view_all")),1)]),_:2},1024)]),n(),C("create zones")&&(((V=c==null?void 0:c.items)==null?void 0:V.length)??0>0)?(a(),y("div",U,[o(i,{appearance:"primary",to:{name:"zone-create-view"}},{default:s(()=>[o(e(N),{size:e(T)},null,8,["size"]),n(" "+l(v("zones.index.create")),1)]),_:2},1024)])):R("",!0)]),n(),o(F,{"data-testid":"zone-control-planes-details",items:c==null?void 0:c.items},null,8,["items"])],64))]}),_:2},1024)]),_:2},1024)):R("",!0),n(),o(h,null,{body:s(()=>[o(_,{src:"/meshes?page=1&size=10"},{default:s(({data:c,error:k})=>[k?(a(),d(w,{key:0,error:k},null,8,["error"])):(a(),y(z,{key:1},[r("div",Y,[r("div",q,[r("h2",null,l(v("main-overview.detail.meshes.title")),1),n(),o(p,{to:{name:"mesh-list-view"}},{default:s(()=>[n(l(v("main-overview.detail.health.view_all")),1)]),_:2},1024)])]),n(),o(D,{"data-testid":"meshes-details",items:c==null?void 0:c.items},null,8,["items"])],64))]),_:2},1024)]),_:2},1024)])])]),_:2},1024)]),_:1})}}});const te=Z(G,[["__scopeId","data-v-8755e20d"]]);export{te as default};
