import{d as v,g as w,r as c,o as _,i as r,w as e,j as l,p as s,n as t,k as n,H as m,l as f,a0 as d,D as g,G as k,a1 as I,t as V}from"./index-a890e85a.js";import{_ as x}from"./EmptyBlock.vue_vue_type_script_setup_true_lang-b1564545.js";const S=i=>(g("data-v-2caa54b3"),i=i(),k(),i),R={class:"summary-title-wrapper"},B=S(()=>s("img",{"aria-hidden":"true",src:I},null,-1)),M={class:"summary-title"},T={key:1,class:"stack"},b={class:"mt-4 stack"},C=v({__name:"MeshSummaryView",props:{name:{},meshInsight:{default:void 0}},setup(i){const{t:a}=w(),o=i;return(N,A)=>{const p=c("RouteTitle"),u=c("RouterLink"),h=c("AppView"),y=c("RouteView");return _(),r(y,{name:"mesh-summary-view"},{default:e(()=>[l(h,null,{title:e(()=>[s("div",R,[B,t(),s("h2",M,[l(u,{to:{name:"mesh-detail-view",params:{mesh:o.name}}},{default:e(()=>[l(p,{title:n(a)("meshes.routes.item.title",{name:o.name}),render:!0},null,8,["title"])]),_:1},8,["to"])])])]),default:e(()=>[t(),o.meshInsight===void 0?(_(),r(x,{key:0},{message:e(()=>[s("p",null,m(n(a)("common.collection.summary.empty_message",{type:"Mesh"})),1)]),default:e(()=>[t(m(n(a)("common.collection.summary.empty_title",{type:"Mesh"}))+" ",1)]),_:1})):(_(),f("div",T,[s("div",null,[s("h3",null,m(n(a)("meshes.routes.item.overview")),1),t(),s("div",b,[l(d,{total:o.meshInsight.services.total??0,"data-testid":"services-status"},{title:e(()=>[t(m(n(a)("meshes.detail.services")),1)]),_:1},8,["total"]),t(),l(d,{online:o.meshInsight.dataplanesByType.standard.online??0,total:o.meshInsight.dataplanesByType.standard.total??0,"data-testid":"data-plane-proxies-status"},{title:e(()=>[t(m(n(a)("meshes.detail.data_plane_proxies")),1)]),_:1},8,["online","total"])])])]))]),_:1})]),_:1})}}});const j=V(C,[["__scopeId","data-v-2caa54b3"]]);export{j as default};
