import{K as N}from"./index-fce48c05.js";import{d as x,k as b,a as c,o as i,b as p,w as e,e as a,f as o,t as m,l as t,q as T,z as A,m as r,c as z,F as S,J as I,p as L,_ as E}from"./index-117d39c8.js";import{E as C}from"./ErrorBlock-0f6b9bc9.js";import{_ as Z}from"./LoadingBlock.vue_vue_type_script_setup_true_lang-2eb92c7c.js";import{A as D}from"./AppCollection-49f203c7.js";import{S as $}from"./StatusBadge-1af5d2b3.js";import"./TextWithCopyButton-9fd2ca95.js";import"./CopyButton-3c068074.js";import"./WarningIcon.vue_vue_type_script_setup_true_lang-729e4f97.js";const F=x({__name:"MeshInsightsList",props:{items:{}},setup(f){const{t:s}=b(),u=f;return(w,k)=>{var d;const y=c("RouterLink");return i(),p(D,{headers:[{label:t(s)("meshes.components.mesh-insights-list.name"),key:"name"},{label:t(s)("meshes.components.mesh-insights-list.services"),key:"services"},{label:t(s)("meshes.components.mesh-insights-list.dataplanes"),key:"dataplanes"}],items:u.items,total:(d=u.items)==null?void 0:d.length,"empty-state-message":t(s)("common.emptyState.message",{type:t(s)("meshes.common.type",{count:2})}),"empty-state-cta-to":t(s)("meshes.href.docs"),"empty-state-cta-text":t(s)("common.documentation")},{name:e(({row:n})=>[a(y,{to:{name:"mesh-detail-view",params:{mesh:n.name}}},{default:e(()=>[o(m(n.name),1)]),_:2},1032,["to"])]),services:e(({row:n})=>[o(m(n.services.internal),1)]),dataplanes:e(({row:n})=>[o(m(n.dataplanesByType.standard.online)+" / "+m(n.dataplanesByType.standard.total),1)]),_:1},8,["headers","items","total","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),J=x({__name:"ZoneControlPlanesList",props:{items:{}},setup(f){const{t:s}=b(),u=T(),w=f;return(k,y)=>{var n;const d=c("RouterLink");return i(),p(D,{headers:[{label:t(s)("zone-cps.components.zone-control-planes-list.name"),key:"name"},{label:t(s)("zone-cps.components.zone-control-planes-list.status"),key:"status"}],items:w.items,total:(n=w.items)==null?void 0:n.length,"empty-state-title":t(s)("zone-cps.empty_state.title"),"empty-state-message":t(u)("create zones")?t(s)("zone-cps.empty_state.message"):t(s)("common.emptyState.message",{type:"Zones"}),"empty-state-cta-to":t(u)("create zones")?{name:"zone-create-view"}:void 0,"empty-state-cta-text":t(s)("zones.index.create")},{name:e(({row:_})=>[a(d,{to:{name:"zone-cp-detail-view",params:{zone:_.name}}},{default:e(()=>[o(m(_.name),1)]),_:2},1032,["to"])]),status:e(({row:_})=>[a($,{status:_.state},null,8,["status"])]),_:1},8,["headers","items","total","empty-state-title","empty-state-message","empty-state-cta-to","empty-state-cta-text"])}}}),q={key:2,class:"stack","data-testid":"detail-view-details"},M={class:"columns"},O={class:"card-header"},U={class:"card-title"},j={key:0,class:"card-actions"},G={class:"card-header"},H={class:"card-title"},Q=x({__name:"ControlPlaneDetailView",setup(f){const s=A();return(u,w)=>{const k=c("RouteTitle"),y=c("RouterLink"),d=c("KButton"),n=c("DataSource"),_=c("KCard"),K=c("AppView"),P=c("RouteView");return i(),p(P,{name:"home"},{default:e(({can:g,t:h})=>[a(K,null,{title:e(()=>[r("h1",null,[a(k,{title:h("main-overview.routes.item.title")},null,8,["title"])])]),default:e(()=>[o(),a(n,{src:"/global-insight"},{default:e(({data:V,error:B})=>[B?(i(),p(C,{key:0,error:B},null,8,["error"])):V===void 0?(i(),p(Z,{key:1})):(i(),z("div",q,[a(t(s),{"can-use-zones":g("use zones"),"global-insight":V},null,8,["can-use-zones","global-insight"]),o(),r("div",M,[g("use zones")?(i(),p(_,{key:0},{default:e(()=>[a(n,{src:"/zone-cps?page=1&size=10"},{default:e(({data:l,error:v})=>{var R;return[v?(i(),p(C,{key:0,error:v},null,8,["error"])):(i(),z(S,{key:1},[r("div",O,[r("div",U,[r("h2",null,m(h("main-overview.detail.zone_control_planes.title")),1),o(),a(y,{to:{name:"zone-cp-list-view"}},{default:e(()=>[o(m(h("main-overview.detail.health.view_all")),1)]),_:2},1024)]),o(),g("create zones")&&(((R=l==null?void 0:l.items)==null?void 0:R.length)??0>0)?(i(),z("div",j,[a(d,{appearance:"primary",to:{name:"zone-create-view"}},{default:e(()=>[a(t(I),{size:t(N)},null,8,["size"]),o(" "+m(h("zones.index.create")),1)]),_:2},1024)])):L("",!0)]),o(),a(J,{"data-testid":"zone-control-planes-details",items:l==null?void 0:l.items},null,8,["items"])],64))]}),_:2},1024)]),_:2},1024)):L("",!0),o(),a(_,null,{default:e(()=>[a(n,{src:"/mesh-insights?page=1&size=10"},{default:e(({data:l,error:v})=>[v?(i(),p(C,{key:0,error:v},null,8,["error"])):(i(),z(S,{key:1},[r("div",G,[r("div",H,[r("h2",null,m(h("main-overview.detail.meshes.title")),1),o(),a(y,{to:{name:"mesh-list-view"}},{default:e(()=>[o(m(h("main-overview.detail.health.view_all")),1)]),_:2},1024)])]),o(),a(F,{"data-testid":"meshes-details",items:l==null?void 0:l.items},null,8,["items"])],64))]),_:2},1024)]),_:2},1024)])]))]),_:2},1024)]),_:2},1024)]),_:1})}}});const ie=E(Q,[["__scopeId","data-v-a360c612"]]);export{ie as default};
