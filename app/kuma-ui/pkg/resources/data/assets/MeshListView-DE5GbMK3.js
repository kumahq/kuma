import{d as z,i as n,o as r,a as d,w as e,j as o,g as _,k as i,t as l,A as h,K as C,e as V,_ as b}from"./index-B4OAi35c.js";import{p as x}from"./kong-icons.es249-CTBvwfdU.js";import{A as L}from"./AppCollection-D_06DYjA.js";import"./kong-icons.es245-uMhO6AtE.js";const D={class:"stack"},R=z({__name:"MeshListView",setup(S){return(A,B)=>{const u=n("RouteTitle"),c=n("RouterLink"),g=n("DataLoader"),y=n("KCard"),f=n("AppView"),w=n("RouteView"),k=n("DataSource");return r(),d(k,{src:"/me"},{default:e(({data:p})=>[p?(r(),d(w,{key:0,name:"mesh-list-view",params:{page:1,size:p.pageSize,mesh:""}},{default:e(({route:a,t})=>[o(f,null,{title:e(()=>[_("h1",null,[o(u,{title:t("meshes.routes.items.title")},null,8,["title"])])]),default:e(()=>[i(),_("div",D,[o(y,null,{default:e(()=>[o(g,{src:`/mesh-insights?page=${a.params.page}&size=${a.params.size}`,loader:!1},{default:e(({data:m,error:v})=>[o(L,{class:"mesh-collection","data-testid":"mesh-collection",headers:[{label:t("meshes.common.name"),key:"name"},{label:t("meshes.routes.items.collection.services"),key:"services"},{label:t("meshes.routes.items.collection.dataplanes"),key:"dataplanes"},{label:"Details",key:"details",hideLabel:!0}],"page-number":a.params.page,"page-size":a.params.size,total:m==null?void 0:m.total,items:m==null?void 0:m.items,error:v,"empty-state-message":t("common.emptyState.message",{type:"Meshes"}),"empty-state-cta-to":t("meshes.href.docs"),"empty-state-cta-text":t("common.documentation"),"is-selected-row":s=>s.name===a.params.mesh,onChange:a.update},{name:e(({row:s})=>[o(c,{to:{name:"mesh-detail-view",params:{mesh:s.name},query:{page:a.params.page,size:a.params.size}}},{default:e(()=>[i(l(s.name),1)]),_:2},1032,["to"])]),services:e(({row:s})=>[i(l(s.services.internal),1)]),dataplanes:e(({row:s})=>[i(l(s.dataplanesByType.standard.online)+" / "+l(s.dataplanesByType.standard.total),1)]),details:e(({row:s})=>[o(c,{class:"details-link","data-testid":"details-link",to:{name:"mesh-detail-view",params:{mesh:s.name}}},{default:e(()=>[i(l(t("common.collection.details_link"))+" ",1),o(h(x),{decorative:"",size:h(C)},null,8,["size"])]),_:2},1032,["to"])]),_:2},1032,["headers","page-number","page-size","total","items","error","empty-state-message","empty-state-cta-to","empty-state-cta-text","is-selected-row","onChange"])]),_:2},1032,["src"])]),_:2},1024)])]),_:2},1024)]),_:2},1032,["params"])):V("",!0)]),_:1})}}}),M=b(R,[["__scopeId","data-v-373101e0"]]);export{M as default};
