import{d as C,g as L,r,o as i,i as _,w as s,j as o,n,H as l,k as e,f as K,l as v,F as f,I as M,Y as O,p as c,E as R,Z as N,K as T,m as b,t as Z}from"./index-28451437.js";import{A as S}from"./AppCollection-b6258a04.js";import{S as A}from"./StatusBadge-fcc954b7.js";import{g as E}from"./index-270afb1d.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-24948e86.js";const D=C({__name:"MeshInsightsList",props:{items:{}},setup(z){const{t}=L(),d=z;return(g,x)=>{var u;const y=r("RouterLink");return i(),_(S,{headers:[{label:e(t)("meshes.components.mesh-insights-list.name"),key:"name"},{label:e(t)("meshes.components.mesh-insights-list.services"),key:"services"},{label:e(t)("meshes.components.mesh-insights-list.dataplanes"),key:"dataplanes"}],items:d.items,total:(u=d.items)==null?void 0:u.length,"empty-state-message":e(t)("common.emptyState.message",{type:e(t)("meshes.common.type",{count:2})}),"empty-state-cta-to":e(t)("meshes.href.docs"),"empty-state-cta-text":e(t)("common.documentation")},{name:s(({row:a})=>[o(y,{to:{name:"mesh-detail-view",params:{mesh:a.name}}},{default:s(()=>[n(l(a.name),1)]),_:2},1032,["to"])]),services:s(({row:a})=>[n(l(a.services.internal??"0"),1)]),dataplanes:s(({row:a})=>[n(l(a.dataplanesByType.standard.online??"0")+" / "+l(a.dataplanesByType.standard.total??"0"),1)]),_:1},8,["headers","items","total","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),F=C({__name:"ZoneControlPlanesList",props:{items:{}},setup(z){const{t}=L(),d=K(),g=z;return(x,y)=>{var a;const u=r("RouterLink");return i(),_(S,{headers:[{label:e(t)("zone-cps.components.zone-control-planes-list.name"),key:"name"},{label:e(t)("zone-cps.components.zone-control-planes-list.status"),key:"status"}],items:g.items,total:(a=g.items)==null?void 0:a.length,"empty-state-title":e(t)("zone-cps.empty_state.title"),"empty-state-message":e(d)("create zones")?e(t)("zone-cps.empty_state.message"):e(t)("common.emptyState.message",{type:"Zones"}),"empty-state-cta-to":e(d)("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":e(t)("zones.index.create")},{name:s(({row:p})=>[o(u,{to:{name:"zone-cp-detail-view",params:{zone:p.name}}},{default:s(()=>[n(l(p.name),1)]),_:2},1032,["to"])]),status:s(({row:p})=>[(i(!0),v(f,null,M([e(E)(p)],w=>(i(),v(f,{key:w},[w?(i(),_(A,{key:0,status:w},null,8,["status"])):(i(),v(f,{key:1},[n(l(e(t)("common.collection.none")),1)],64))],64))),128))]),_:1},8,["headers","items","total","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),P={class:"stack","data-testid":"detail-view-details"},$={class:"columns"},j={class:"card-header"},H={class:"card-title"},U={key:0,class:"card-actions"},Y={class:"card-header"},q={class:"card-title"},G=C({__name:"MainOverviewView",setup(z){const t=O();return(d,g)=>{const x=r("RouteTitle"),y=r("RouterLink"),u=r("KButton"),a=r("DataSource"),p=r("KCard"),w=r("AppView"),I=r("RouteView");return i(),_(I,{name:"home"},{default:s(({can:V,t:h})=>[o(w,null,{title:s(()=>[c("h1",null,[o(x,{title:h("main-overview.routes.item.title"),render:!0},null,8,["title"])])]),default:s(()=>[n(),c("div",P,[o(e(t)),n(),c("div",$,[V("use zones")?(i(),_(p,{key:0},{body:s(()=>[o(a,{src:"/zone-cps?page=1&size=10"},{default:s(({data:m,error:k})=>{var B;return[k?(i(),_(R,{key:0,error:k},null,8,["error"])):(i(),v(f,{key:1},[c("div",j,[c("div",H,[c("h2",null,l(h("main-overview.detail.zone_control_planes.title")),1),n(),o(y,{to:{name:"zone-cp-list-view"}},{default:s(()=>[n(l(h("main-overview.detail.health.view_all")),1)]),_:2},1024)]),n(),V("create zones")&&(((B=m==null?void 0:m.items)==null?void 0:B.length)??0>0)?(i(),v("div",U,[o(u,{appearance:"primary",to:{name:"zone-create-view"}},{default:s(()=>[o(e(N),{size:e(T)},null,8,["size"]),n(" "+l(h("zones.index.create")),1)]),_:2},1024)])):b("",!0)]),n(),o(F,{"data-testid":"zone-control-planes-details",items:m==null?void 0:m.items},null,8,["items"])],64))]}),_:2},1024)]),_:2},1024)):b("",!0),n(),o(p,null,{body:s(()=>[o(a,{src:"/meshes?page=1&size=10"},{default:s(({data:m,error:k})=>[k?(i(),_(R,{key:0,error:k},null,8,["error"])):(i(),v(f,{key:1},[c("div",Y,[c("div",q,[c("h2",null,l(h("main-overview.detail.meshes.title")),1),n(),o(y,{to:{name:"mesh-list-view"}},{default:s(()=>[n(l(h("main-overview.detail.health.view_all")),1)]),_:2},1024)])]),n(),o(D,{"data-testid":"meshes-details",items:m==null?void 0:m.items},null,8,["items"])],64))]),_:2},1024)]),_:2},1024)])])]),_:2},1024)]),_:1})}}});const te=Z(G,[["__scopeId","data-v-def7344b"]]);export{te as default};
